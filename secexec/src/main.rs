use std::{
    error::Error,
    ffi::{CString, NulError},
};

use clap::{crate_authors, crate_name, crate_version, App, Arg};

use secexec::{Config, Executor};

/// For security reasons:
///  - `POLICY_PATH` must be read-only.
fn main() -> Result<(), Box<dyn Error>> {
    let matches = App::new(crate_name!())
        .version(crate_version!())
        .author(crate_authors!())
        .usage("secexec <POLICY_PATH> --pwd <PATH> -- <EXEC> [<ARGS>...]")
        .arg(
            Arg::with_name("POLICY_PATH")
                .help("Path with seccomp policies")
                .required(true)
                .index(1),
        )
        .arg(
            Arg::with_name("ARGS")
                .help("Executable arguments (including executable path)")
                .required(true)
                .multiple(true),
        )
        .arg(
            Arg::with_name("pwd")
                .long("pwd")
                .value_name("PATH")
                .default_value("/")
                .help("Current working directory")
                .takes_value(true),
        )
        .get_matches();

    // This is safe, because `POLICY_PATH` argument is required.
    let policy_path = matches.value_of("POLICY_PATH").unwrap();
    // This is safe, because these arguments have default value.
    let pwd = matches.value_of("pwd").unwrap();

    // Convert trailing arguments to a vector of `CString`.
    let args = matches
        .values_of("ARGS")
        .unwrap_or_default()
        .map(|v| CString::new(v))
        .collect::<Result<Vec<CString>, NulError>>()?;

    // This is safe, because `ARGS` argument is required and has at least a single value.
    let exec = args.first().unwrap();

    let cfg = Config::load(policy_path, exec.to_str()?)?;

    unsafe { libc::chdir(CString::new(pwd)?.as_ptr()) };

    Executor::new(cfg).exec(exec, &args, &[])?;

    Ok(())
}
