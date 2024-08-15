use crate::api::{PsiExecuteResult, RequestHeader, ResultHeader};
use anyhow::Error;
use axum::response::{IntoResponse, Response};
use std::collections::HashMap;

#[derive(Debug)]
pub struct AppError {
    pub err: Error,
}

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        PsiExecuteResult {
            header: Some(ResultHeader {
                request_id: "".to_string(),
                metadata: HashMap::new(),
                code: -1,
                msg: format!("err:{}", self.err),
            }),
            keys: Vec::new(),
        }
        .into_response()
    }
}

impl AppError {
    pub fn into_tonic_response(
        self,
        header: &Option<RequestHeader>,
    ) -> tonic::Response<PsiExecuteResult> {
        if header.is_none() {
            tonic::Response::new(PsiExecuteResult {
                header: Some(ResultHeader {
                    request_id: "".to_string(),
                    metadata: HashMap::new(),
                    code: -1,
                    msg: format!("err:{}", self.err),
                }),
                keys: Vec::new(),
            })
        } else {
            let value = header.as_ref().unwrap();
            tonic::Response::new(PsiExecuteResult {
                header: Some(ResultHeader {
                    request_id: value.request_id.clone(),
                    metadata: value.metadata.clone(),
                    code: -1,
                    msg: format!("err:{}", self.err),
                }),
                keys: Vec::new(),
            })
        }
    }
}

impl From<AppError> for PsiExecuteResult {
    fn from(err: AppError) -> Self {
        PsiExecuteResult {
            header: Some(ResultHeader {
                request_id: "".to_string(),
                metadata: HashMap::new(),
                code: -1,
                msg: format!("err:{}", err.err),
            }),
            keys: Vec::new(),
        }
    }
}

impl<E> From<E> for AppError
where
    E: Into<anyhow::Error>,
{
    fn from(err: E) -> Self {
        Self { err: err.into() }
    }
}
