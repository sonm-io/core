use std::io::Error;

/// Set the `NO_NEW_PRIVS` attribute to the current thread.
///
/// Once set, this the `NO_NEW_PRIVS` attribute cannot be unset.
/// The setting of this attribute is inherited by children created
/// by `fork(2)` and `clone(2)`, and preserved across `execve(2)`.
#[inline]
pub fn set_no_new_privs() -> Result<(), Error> {
    if unsafe { libc::prctl(libc::PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0) } == 0 {
        Ok(())
    } else {
        Err(Error::last_os_error())
    }
}

/// Set the state of the "dumpable" flag, which determines whether core dumps
/// are produced for the calling process upon delivery of a signal whose
/// default behavior is to produce a core dump.
///
/// Processes that are not dumpable can not be attached
/// via `ptrace(2)` `PTRACE_ATTACH`; see `ptrace(2)` for further details.
#[inline]
pub fn set_dumpable(dumpable: bool) -> Result<(), Error> {
    let v = if dumpable {
        1
    } else {
        0
    };

    if unsafe { libc::prctl(libc::PR_SET_DUMPABLE, v) } == 0 {
        Ok(())
    } else {
        Err(Error::last_os_error())
    }
}
