//! DeliveryService trait implementation
//!
//! This module contains the tonic gRPC trait implementation.
//! Each method delegates to the appropriate handler in `handlers/`.

use tonic::{Request, Response, Status};
use tracing::instrument;

use super::DeliveryServiceImpl;
use crate::infrastructure::rpc::handlers::{courier, order};
use crate::infrastructure::rpc::{
    AcceptOrderRequest, AcceptOrderResponse, ActivateCourierRequest, ActivateCourierResponse,
    ArchiveCourierRequest, ArchiveCourierResponse, AssignOrderRequest, AssignOrderResponse,
    ChangeTransportTypeRequest, ChangeTransportTypeResponse, DeactivateCourierRequest,
    DeactivateCourierResponse, DeliveryService, GetCourierPoolRequest, GetCourierPoolResponse,
    GetCourierRequest, GetCourierResponse, RegisterCourierRequest, RegisterCourierResponse,
    UpdateContactInfoRequest, UpdateContactInfoResponse, UpdateWorkScheduleRequest,
    UpdateWorkScheduleResponse,
};

#[tonic::async_trait]
impl DeliveryService for DeliveryServiceImpl {
    // ==================== Courier Queries ====================

    #[instrument(skip(self, request), fields(courier_id = %request.get_ref().courier_id))]
    async fn get_courier(
        &self,
        request: Request<GetCourierRequest>,
    ) -> Result<Response<GetCourierResponse>, Status> {
        courier::get_courier(&self.state, request.into_inner()).await
    }

    #[instrument(skip(self, request))]
    async fn get_courier_pool(
        &self,
        request: Request<GetCourierPoolRequest>,
    ) -> Result<Response<GetCourierPoolResponse>, Status> {
        courier::get_courier_pool(&self.state, request.into_inner()).await
    }

    // ==================== Courier Lifecycle ====================

    #[instrument(skip(self, request), fields(email = %request.get_ref().email))]
    async fn register_courier(
        &self,
        request: Request<RegisterCourierRequest>,
    ) -> Result<Response<RegisterCourierResponse>, Status> {
        courier::register_courier(&self.state, request.into_inner()).await
    }

    #[instrument(skip(self, request), fields(courier_id = %request.get_ref().courier_id))]
    async fn activate_courier(
        &self,
        request: Request<ActivateCourierRequest>,
    ) -> Result<Response<ActivateCourierResponse>, Status> {
        courier::activate_courier(&self.state, request.into_inner()).await
    }

    #[instrument(skip(self, request), fields(courier_id = %request.get_ref().courier_id))]
    async fn deactivate_courier(
        &self,
        request: Request<DeactivateCourierRequest>,
    ) -> Result<Response<DeactivateCourierResponse>, Status> {
        courier::deactivate_courier(&self.state, request.into_inner()).await
    }

    #[instrument(skip(self, request), fields(courier_id = %request.get_ref().courier_id))]
    async fn archive_courier(
        &self,
        request: Request<ArchiveCourierRequest>,
    ) -> Result<Response<ArchiveCourierResponse>, Status> {
        courier::archive_courier(&self.state, request.into_inner()).await
    }

    // ==================== Courier Profile ====================

    #[instrument(skip(self, request), fields(courier_id = %request.get_ref().courier_id))]
    async fn update_contact_info(
        &self,
        request: Request<UpdateContactInfoRequest>,
    ) -> Result<Response<UpdateContactInfoResponse>, Status> {
        courier::update_contact_info(&self.state, request.into_inner()).await
    }

    #[instrument(skip(self, request), fields(courier_id = %request.get_ref().courier_id))]
    async fn update_work_schedule(
        &self,
        request: Request<UpdateWorkScheduleRequest>,
    ) -> Result<Response<UpdateWorkScheduleResponse>, Status> {
        courier::update_work_schedule(&self.state, request.into_inner()).await
    }

    #[instrument(skip(self, request), fields(courier_id = %request.get_ref().courier_id))]
    async fn change_transport_type(
        &self,
        request: Request<ChangeTransportTypeRequest>,
    ) -> Result<Response<ChangeTransportTypeResponse>, Status> {
        courier::change_transport_type(&self.state, request.into_inner()).await
    }

    // ==================== Order Operations ====================

    #[instrument(skip(self, request), fields(order_id = %request.get_ref().order_id))]
    async fn accept_order(
        &self,
        request: Request<AcceptOrderRequest>,
    ) -> Result<Response<AcceptOrderResponse>, Status> {
        order::accept_order(&self.state, request.into_inner()).await
    }

    #[instrument(skip(self, request), fields(package_id = %request.get_ref().package_id))]
    async fn assign_order(
        &self,
        request: Request<AssignOrderRequest>,
    ) -> Result<Response<AssignOrderResponse>, Status> {
        order::assign_order(&self.state, request.into_inner()).await
    }
}
