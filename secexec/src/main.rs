use std::{
    error::Error,
    ffi::{CStr, CString, NulError},
};

use secexec::{Config, Executor};

/// Poor man command-line arguments parser.
struct Args {
    /// Path to a directory with seccomp policies.
    path: String,
    /// Executable arguments (including executable path, i.e. what to execute).
    argv: Vec<CString>,
}

impl Args {
    /// Loads program arguments and extracts the required.
    pub fn new() -> Result<Self, Box<dyn Error>> {
        // Skip our executable name.
        let mut argv = std::env::args().skip(1);

        let path = match argv.next() {
            Some(path) => path,
            None => return Err("missing required policy path argument".into()),
        };

        let argv: Result<Vec<CString>, NulError> = argv.map(|v| CString::new(v)).collect();
        let argv = argv?;

        if argv.is_empty() {
            return Err("missing executable".into());
        }

        let m = Self { path, argv };

        Ok(m)
    }

    /// Returns path containing policy rules.
    #[inline]
    pub fn policy_path(&self) -> &str {
        &self.path
    }

    /// Returns the executable path.
    #[inline]
    pub fn exec(&self) -> &CStr {
        &self.argv[0]
    }

    /// Returns executable arguments including full program name.
    #[inline]
    pub fn argv(&self) -> &[CString] {
        &self.argv
    }
}

/// Usage: `secexec POLICY_PATH EXEC_PATH [ARGS...]`.
///
/// For security reasons:
///  - `POLICY_PATH` must be read-only.
fn main() -> Result<(), Box<dyn Error>> {
    let args = Args::new()?;
    let exec = args.exec().to_str()?;
    let cfg = Config::load(args.policy_path(), exec)?;

    // Change CWD to "/" because why not.
    unsafe { libc::chdir(CString::new("/")?.as_ptr()) };

    Executor::new(cfg).exec(args.exec(), args.argv(), &[])?;

    Ok(())
}
