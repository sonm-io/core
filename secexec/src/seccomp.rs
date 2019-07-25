use std::{
    error,
    ffi::CStr,
    fmt::{self, Display, Formatter},
    io::Error,
};

use serde::Deserialize;

/// Comparison operator.
#[derive(Debug, Copy, Clone, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum Op {
    /// Strict equality.
    Eq,
    /// Matches when the masked argument value is equal to the masked datum
    /// value.
    MaskedEq,
}

impl Op {
    /// Returns this operator value in a form of seccomp internal structure.
    pub fn as_compare(&self) -> seccomp_sys::scmp_compare {
        match self {
            Op::Eq => seccomp_sys::scmp_compare::SCMP_CMP_EQ,
            Op::MaskedEq => seccomp_sys::scmp_compare::SCMP_CMP_MASKED_EQ,
        }
    }
}

/// Argument value comparison.
#[derive(Debug, Copy, Clone, Deserialize)]
pub struct ArgCmp {
    /// Argument position.
    arg: u32,
    /// Comparison operator.
    op: Op,
    /// Comparison right hand value.
    a: u64,
    /// Used to compare masked value when using `SCMP_CMP_MASKED_EQ`.
    #[serde(default)]
    b: u64,
}

impl ArgCmp {
    pub fn new(arg: u32, op: Op, a: u64, b: u64) -> Self {
        Self { arg, op, a, b }
    }

    /// Returns the seccomp argument comparison value for this rule.
    pub fn as_arg_cmp(&self) -> seccomp_sys::scmp_arg_cmp {
        seccomp_sys::scmp_arg_cmp {
            arg: self.arg,
            op: self.op.as_compare(),
            datum_a: self.a,
            datum_b: self.b,
        }
    }
}

/// Seccomp actions.
#[derive(Debug, Clone, Copy)]
pub enum Action {
    /// Allow the syscall to be executed
    Allow,
    /// Kill the process
    Kill,
    /// Throw a SIGSYS signal
    Trap,
    /// Return the specified error code
    Errno(i32),
    /// Notify a tracing process with the specified value
    Trace(u32),
}

impl Into<u32> for Action {
    fn into(self) -> u32 {
        match self {
            Action::Allow => seccomp_sys::SCMP_ACT_ALLOW,
            Action::Kill => seccomp_sys::SCMP_ACT_KILL,
            Action::Trap => seccomp_sys::SCMP_ACT_TRAP,
            Action::Errno(v) => seccomp_sys::SCMP_ACT_ERRNO(v as u32),
            Action::Trace(v) => seccomp_sys::SCMP_ACT_TRACE(v),
        }
    }
}

#[derive(Debug)]
pub struct SeccompInitFailed;

impl Display for SeccompInitFailed {
    fn fmt(&self, fmt: &mut Formatter) -> Result<(), fmt::Error> {
        fmt.write_str("seccomp initialization failed")
    }
}

impl error::Error for SeccompInitFailed {}

#[derive(Debug)]
pub struct Context {
    ctx: *mut seccomp_sys::scmp_filter_ctx,
}

impl Context {
    /// Constructs a new seccomp context.
    pub fn new(action: Action) -> Result<Self, SeccompInitFailed> {
        match unsafe { seccomp_sys::seccomp_init(action.into()) } {
            ctx if ctx.is_null() => Err(SeccompInitFailed),
            ctx => Ok(Self { ctx }),
        }
    }

    /// Adds a seccomp filter rule.
    pub fn add_rule(
        &mut self,
        syscall: &CStr,
        action: Action,
        cmp: Option<ArgCmp>,
    ) -> Result<(), Error> {
        let id = unsafe { seccomp_sys::seccomp_syscall_resolve_name(syscall.as_ptr()) };

        let rc = match cmp {
            Some(rule) => unsafe {
                seccomp_sys::seccomp_rule_add(self.ctx, action.into(), id, 1, rule.as_arg_cmp())
            },
            None => unsafe { seccomp_sys::seccomp_rule_add(self.ctx, action.into(), id, 0) },
        };

        if rc == 0 {
            Ok(())
        } else {
            return Err(Error::from_raw_os_error(-rc));
        }
    }

    /// Loads the current seccomp filter into the kernel.
    pub fn load(self) -> Result<(), Error> {
        match unsafe { seccomp_sys::seccomp_load(self.ctx) } {
            0 => Ok(()),
            e => Err(Error::from_raw_os_error(-e)),
        }
    }
}

impl Drop for Context {
    fn drop(&mut self) {
        unsafe { seccomp_sys::seccomp_release(self.ctx) }
    }
}
