use core::str::FromStr;
use std::{
    collections::{HashMap, HashSet},
    error::Error,
    ffi::CString,
    fs::File,
    path::Path,
};

use caps::Capability;
use serde::Deserialize;

use crate::seccomp::ArgCmp;

#[derive(Debug, Deserialize)]
pub struct Config {
    /// Executable checksum.
    checksum: String,
    /// Capabilities whitelist.
    capabilities: Option<HashSet<String>>,
    /// Syscall whitelist.
    syscalls: HashMap<CString, Vec<ArgCmp>>,
}

impl Config {
    /// Loads the security config from the specified path for the specified
    /// executable.
    pub fn load<P: AsRef<Path>>(path: P, name: &str) -> Result<Self, Box<dyn Error>> {
        let path = match Path::new(name).file_stem() {
            Some(name) => {
                path.as_ref().join(name).with_extension("yaml")
            }
            None => {
                return Err("invalid executable file name".into());
            }
        };

        let file = File::open(&path)?; // todo: human readable errors.
        let cfg = serde_yaml::from_reader(file)?; // todo: human readable errors.

        Ok(cfg)
    }

    pub fn sha512(&self) -> &str {
        &self.checksum
    }

    pub fn syscalls(&self) -> HashMap<CString, Vec<ArgCmp>> {
        self.syscalls.clone()
    }

    pub fn capabilities(&self) -> Result<HashSet<caps::Capability>, Box<dyn Error>> {
        match &self.capabilities {
            Some(capabilities) => {
                let mut caps = HashSet::new();
                for cap in capabilities {
                    caps.insert(Capability::from_str(&cap.to_uppercase())?);
                }

                Ok(caps)
            }
            None => Ok(HashSet::new())
        }
    }
}
