use crate::api::{PsiExecuteRequest, PsiExecuteResult, RequestHeader, ResultHeader};
use crate::front::err::AppError;
use crate::policy::new;
use prost_types::Any;
use std::collections::HashMap;
use std::hash::Hash;

use axum::{
    async_trait,
    body::Bytes,
    extract::{FromRequest, Request},
    response::{IntoResponse, Response},
};
use prost::Message;

#[async_trait]
impl<S> FromRequest<S> for PsiExecuteRequest
where
    S: Send + Sync,
{
    type Rejection = AppError;

    async fn from_request(req: Request, state: &S) -> Result<Self, Self::Rejection> {
        let body = Bytes::from_request(req, state).await?;

        let req: PsiExecuteRequest = PsiExecuteRequest::decode(body)?;

        Ok(req)
    }
}

impl IntoResponse for PsiExecuteResult {
    fn into_response(self) -> Response {
        self.encode_to_vec().into_response()
    }
}

pub fn back_header(header: &Option<RequestHeader>) -> Option<ResultHeader> {
    if header.is_none() {
        Some(ResultHeader {
            request_id: "".to_string(),
            metadata: HashMap::<String, Any>::new(),
            code: 0,
            msg: "".to_string(),
        })
    } else {
        let value = header.as_ref().unwrap();
        Some(ResultHeader {
            request_id: value.request_id.clone(),
            metadata: value.metadata.clone(),
            code: 0,
            msg: "".to_string(),
        })
    }
}
