pub mod cover;
pub mod err;

use crate::api::PsiExecuteRequest;
use crate::execute::ExecuteEngine;
use crate::front::err::AppError;
use axum::{
    extract::State,
    response::{IntoResponse, Response},
};
use std::sync::Arc;

#[derive(Clone)]
pub struct AppStateDyn {
    pub engine: Arc<Box<ExecuteEngine>>,
}

pub async fn do_psi(
    state: State<AppStateDyn>,
    req: PsiExecuteRequest,
) -> Result<Response, AppError> {
    do_psi_impl(state, &req).await.or_else(|err| {
        Ok(err
            .into_tonic_response(&req.header)
            .get_ref()
            .clone()
            .into_response())
    })
}

pub async fn do_psi_impl(
    state: State<AppStateDyn>,
    req: &PsiExecuteRequest,
) -> Result<Response, AppError> {
    let rsp = state.engine.client.psi_execute(&req).await?;

    {
        let request_info = rsp.header.clone().unwrap();
        tracing::info!(
            "request_id[{}] code:[{}] msg:[{}]",
            request_info.request_id,
            request_info.code,
            request_info.msg,
        );
    }

    Ok(rsp.into_response())
}
