//! Secured execution of arbitrary commands.
//!
//! This library contains functionality that allows executing processes in a
//! jailed environment, restricting their capabilities and allowed syscalls,
//! to prevent leaking sensitive information.

pub use crate::{cfg::Config, exec::Executor};

mod cfg;
mod exec;
pub mod prctl;
pub mod seccomp;
