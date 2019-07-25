use std::{
    collections::HashMap,
    error::Error,
    ffi::{CStr, CString},
    fs::File,
    path::Path,
};

use sha2::{digest::Digest, Sha512};

use crate::{
    cfg::Config,
    prctl,
    seccomp::{Action, ArgCmp, Context, Op},
};

fn init_seccomp(syscalls: &HashMap<CString, Vec<ArgCmp>>) -> Result<(), Box<dyn Error>> {
    // Disable extending capabilities for child processes if any.
    prctl::set_no_new_privs()?;
    // Forbid attempts to attach a ptracer.
    prctl::set_dumpable(false)?;

    // Forbidden syscalls will return EPERM. It is sometimes better than to
    // segfault, because many applications still work properly.
    let mut ctx = Context::new(Action::Errno(libc::EPERM))?;

    for (syscall, rules) in syscalls {
        if rules.is_empty() {
            ctx.add_rule(syscall, Action::Allow, None)?;
        } else {
            for rule in rules {
                ctx.add_rule(syscall, Action::Allow, Some(*rule))?;
            }
        }
    }

    ctx.load()?;

    //    println!("[!] PREPARE YOUR ANUS");

    Ok(())
}

fn to_exec_array(args: &[CString]) -> Vec<*const libc::c_char> {
    let mut args: Vec<*const libc::c_char> = args.iter().map(|s| s.as_ptr()).collect();
    args.push(core::ptr::null());
    args
}

#[derive(Debug)]
pub struct Executor {
    cfg: Config,
}

impl Executor {
    pub fn new(cfg: Config) -> Self {
        Self { cfg }
    }

    pub fn exec(
        &self,
        path: &CStr,
        args: &[CString],
        envp: &[CString],
    ) -> Result<(), Box<dyn Error>> {
        self.verify_checksum(path)?;
        self.drop_capabilities()?;

        let args_p = to_exec_array(args);
        let envp_p = to_exec_array(envp);

        init_seccomp(&self.allowed_syscalls(path))?;
        unsafe { libc::execve(path.as_ptr(), args_p.as_ptr(), envp_p.as_ptr()) };

        Err(std::io::Error::last_os_error().into())
    }

    fn verify_checksum(&self, path: &CStr) -> Result<(), Box<dyn Error>> {
        let mut file = File::open(Path::new(path.to_str()?))?;
        let mut hash = Sha512::new();
        std::io::copy(&mut file, &mut hash)?;

        if format!("{:x}", hash.result()) != self.cfg.sha512() {
            return Err("invalid executable checksum".into());
        }

        Ok(())
    }

    fn drop_capabilities(&self) -> Result<(), Box<dyn Error>> {
        for cap in caps::all().difference(&self.cfg.capabilities()?) {
            //            println!("[-] Dropping {}", cap);
            caps::drop(None, caps::CapSet::Ambient, *cap)?;
        }

        Ok(())
    }

    fn allowed_syscalls(&self, path: &CStr) -> HashMap<CString, Vec<ArgCmp>> {
        let mut syscalls = self.cfg.syscalls();
        syscalls.insert(
            CString::new("execve").unwrap(),
            vec![ArgCmp::new(0, Op::Eq, path.as_ptr() as u64, 0)],
        );
        syscalls.insert(CString::new("exit_group").unwrap(), Vec::new());
        syscalls.insert(CString::new("exit").unwrap(), Vec::new());

        syscalls
    }
}
