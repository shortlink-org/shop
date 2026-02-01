//! Courier Commands
//!
//! Commands that modify courier state.

// Lifecycle commands
pub mod activate;
pub mod archive;
pub mod deactivate;
pub mod register;

// Profile commands
pub mod change_transport_type;
pub mod update_contact_info;
pub mod update_location;
pub mod update_work_schedule;

// Re-export main types - Lifecycle
pub use activate::{
    ActivateCourierError, Command as ActivateCommand, Handler as ActivateHandler,
    Response as ActivateResponse,
};
pub use archive::{
    ArchiveCourierError, Command as ArchiveCommand, Handler as ArchiveHandler,
    Response as ArchiveResponse,
};
pub use deactivate::{
    Command as DeactivateCommand, DeactivateCourierError, Handler as DeactivateHandler,
    Response as DeactivateResponse,
};
pub use register::{
    Command as RegisterCommand, Handler as RegisterHandler, RegisterCourierError,
    Response as RegisterResponse,
};

// Re-export main types - Profile
pub use change_transport_type::{
    ChangeTransportTypeError, Command as ChangeTransportTypeCommand,
    Handler as ChangeTransportTypeHandler, Response as ChangeTransportTypeResponse,
};
pub use update_contact_info::{
    Command as UpdateContactInfoCommand, Handler as UpdateContactInfoHandler,
    Response as UpdateContactInfoResponse, UpdateContactInfoError,
};
pub use update_location::{
    Command as UpdateLocationCommand, Handler as UpdateLocationHandler, UpdateCourierLocationError,
};
pub use update_work_schedule::{
    Command as UpdateWorkScheduleCommand, Handler as UpdateWorkScheduleHandler,
    Response as UpdateWorkScheduleResponse, UpdateWorkScheduleError,
};
