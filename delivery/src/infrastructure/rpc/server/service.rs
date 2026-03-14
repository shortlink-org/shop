//! DeliveryService trait implementation
//!
//! This module contains the tonic gRPC trait implementation.
//! Each method delegates to the appropriate handler in `handlers/`.

use std::pin::Pin;
use std::str::FromStr;

use futures::Stream;
use futures_util::StreamExt;
use tokio::sync::mpsc;
use tokio_stream::wrappers::ReceiverStream;
use tonic::{Request, Response, Status};
use tracing::instrument;
use uuid::Uuid;

use super::DeliveryServiceImpl;
use crate::infrastructure::rpc::handlers::{courier, order, random_address, tracking};
use crate::infrastructure::rpc::{
    AcceptOrderRequest, AcceptOrderResponse, ActivateCourierRequest, ActivateCourierResponse,
    ArchiveCourierRequest, ArchiveCourierResponse, AssignOrderRequest, AssignOrderResponse,
    ChangeTransportTypeRequest, ChangeTransportTypeResponse, DeactivateCourierRequest,
    DeactivateCourierResponse, DeliverOrderRequest, DeliverOrderResponse, DeliveryService,
    GetCourierDeliveriesRequest, GetCourierDeliveriesResponse, GetCourierPoolRequest,
    GetCourierPoolResponse, GetCourierRequest, GetCourierResponse, GetOrderTrackingRequest,
    GetOrderTrackingResponse, GetRandomAddressRequest, GetRandomAddressResponse,
    PickUpOrderRequest, PickUpOrderResponse, RegisterCourierRequest, RegisterCourierResponse,
    UpdateContactInfoRequest, UpdateContactInfoResponse, UpdateWorkScheduleRequest,
    UpdateWorkScheduleResponse,
};

#[tonic::async_trait]
impl DeliveryService for DeliveryServiceImpl {
    type SubscribeOrderTrackingStream =
        Pin<Box<dyn Stream<Item = Result<GetOrderTrackingResponse, Status>> + Send + 'static>>;

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

    #[instrument(skip(self, request), fields(package_id = %request.get_ref().package_id, courier_id = %request.get_ref().courier_id))]
    async fn pick_up_order(
        &self,
        request: Request<PickUpOrderRequest>,
    ) -> Result<Response<PickUpOrderResponse>, Status> {
        order::pick_up_order(&self.state, request.into_inner()).await
    }

    #[instrument(skip(self, request), fields(package_id = %request.get_ref().package_id, courier_id = %request.get_ref().courier_id))]
    async fn deliver_order(
        &self,
        request: Request<DeliverOrderRequest>,
    ) -> Result<Response<DeliverOrderResponse>, Status> {
        order::deliver_order(&self.state, request.into_inner()).await
    }

    // ==================== Courier Deliveries ====================

    #[instrument(skip(self, request), fields(courier_id = %request.get_ref().courier_id))]
    async fn get_courier_deliveries(
        &self,
        request: Request<GetCourierDeliveriesRequest>,
    ) -> Result<Response<GetCourierDeliveriesResponse>, Status> {
        courier::get_courier_deliveries(&self.state, request.into_inner()).await
    }

    #[instrument(skip(self, request), fields(order_id = %request.get_ref().order_id))]
    async fn get_order_tracking(
        &self,
        request: Request<GetOrderTrackingRequest>,
    ) -> Result<Response<GetOrderTrackingResponse>, Status> {
        let customer_id = customer_id_from_request(&request)?;
        tracking::get_order_tracking(&self.state, request.into_inner(), customer_id).await
    }

    #[instrument(skip(self, request), fields(order_id = %request.get_ref().order_id))]
    async fn subscribe_order_tracking(
        &self,
        request: Request<GetOrderTrackingRequest>,
    ) -> Result<Response<Self::SubscribeOrderTrackingStream>, Status> {
        let customer_id = customer_id_from_request(&request)?;
        let order_id = Uuid::from_str(&request.get_ref().order_id)
            .map_err(|_| Status::invalid_argument("invalid order_id format"))?;
        let initial =
            tracking::build_order_tracking_response(&self.state, order_id, customer_id).await?;
        let mut tracking_updates = self
            .state
            .tracking_pubsub
            .subscribe_order_updates(order_id)
            .await
            .map_err(|err| Status::internal(err.to_string()))?;

        let state = self.state.clone();
        let (tx, rx) = mpsc::channel(16);

        tokio::spawn(async move {
            if tx.send(Ok(initial)).await.is_err() {
                return;
            }

            loop {
                tokio::select! {
                    _ = tx.closed() => return,
                    message = tracking_updates.next() => {
                        if message.is_none() {
                            return;
                        }

                        match tracking::build_order_tracking_response(&state, order_id, customer_id).await {
                            Ok(snapshot) => {
                                if tx.send(Ok(snapshot)).await.is_err() {
                                    return;
                                }
                            }
                            Err(err) => {
                                if tx.send(Err(err)).await.is_err() {
                                    return;
                                }
                            }
                        }
                    }
                }
            }
        });

        Ok(Response::new(
            Box::pin(ReceiverStream::new(rx)) as Self::SubscribeOrderTrackingStream
        ))
    }

    // ==================== Random Address ====================

    #[instrument(skip(self, request))]
    async fn get_random_address(
        &self,
        request: Request<GetRandomAddressRequest>,
    ) -> Result<Response<GetRandomAddressResponse>, Status> {
        random_address::get_random_address(&self.state, request.into_inner()).await
    }
}

fn customer_id_from_request<T>(request: &Request<T>) -> Result<Uuid, Status> {
    let metadata = request.metadata();
    let raw = metadata
        .get("x-user-id")
        .ok_or_else(|| Status::unauthenticated("missing x-user-id metadata"))?
        .to_str()
        .map_err(|_| Status::unauthenticated("invalid x-user-id metadata"))?;

    Uuid::from_str(raw).map_err(|_| Status::unauthenticated("invalid x-user-id metadata"))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn customer_id_from_request_reads_uuid_metadata() {
        let customer_id = Uuid::new_v4();
        let mut request = Request::new(GetOrderTrackingRequest {
            order_id: Uuid::new_v4().to_string(),
        });
        request
            .metadata_mut()
            .insert("x-user-id", customer_id.to_string().parse().unwrap());

        let parsed = customer_id_from_request(&request).unwrap();
        assert_eq!(parsed, customer_id);
    }

    #[test]
    fn customer_id_from_request_requires_metadata() {
        let request = Request::new(GetOrderTrackingRequest {
            order_id: Uuid::new_v4().to_string(),
        });

        let err = customer_id_from_request(&request).unwrap_err();
        assert_eq!(err.code(), tonic::Code::Unauthenticated);
    }
}
