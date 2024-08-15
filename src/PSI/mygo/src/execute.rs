// Copyright 2020 The 9nFL Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

pub mod client;
pub mod serve;

use crate::api::execute_service_server::ExecuteServiceServer;
use crate::encrypt::Curve;
use crate::execute::client::Client;
use crate::execute::serve::ExecuteServiceImpl;
use crate::front::err::AppError;
use hyper::service::make_service_fn;
use local_ip_address::local_ip;
use redis::Commands;
use std::convert::Infallible;
use std::fmt::Debug;
use std::net::SocketAddr;
use std::str::FromStr;
use tokio::spawn;
use tokio::sync::oneshot;
use tokio::time::{sleep, Duration};

pub struct ExecuteEngine {
    pub client: Client,
    pub join_handle: tokio::task::JoinHandle<()>,
    pub shutdown: oneshot::Sender<()>,
}

impl Debug for ExecuteEngine {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("ExecuteEngine").finish()
    }
}

impl ExecuteEngine {
    pub async fn new(
        key: String,
        curve: String,
        host: String,
        remote: String,
        target: String,
        id: String,
        redis_address: String,
        redis_password: String,
    ) -> Result<Self, AppError> {
        //for debug
        if !redis_address.is_empty() {
            let redis_url = {
                if redis_password.is_empty() {
                    format!("redis://{redis_address}/")
                } else {
                    format!("redis://:{redis_password}@{redis_address}")
                }
            };

            redis::Client::open(redis_url)?.get_connection()?.set(
                format!("network:{}", id),
                format!(
                    "{}:{}",
                    local_ip()?.to_string(),
                    SocketAddr::from_str(&host)?.port()
                ),
            )?;
        }

        let (shutdown, rx) = oneshot::channel::<()>();
        let server_curve = Curve::new(key.as_bytes(), &curve);
        let join_handle = spawn(async move {
            let grpc_service = tonic::transport::Server::builder()
                .add_service(ExecuteServiceServer::new(ExecuteServiceImpl::new(
                    server_curve,
                )))
                .into_service();
            let builder = hyper::Server::bind(&SocketAddr::from_str(&host).unwrap());

            let make_grpc_service = make_service_fn(move |_conn| {
                let grpc_service = grpc_service.clone();
                async { Ok::<_, Infallible>(grpc_service) }
            });

            builder
                .serve(make_grpc_service)
                .with_graceful_shutdown(async {
                    rx.await.ok();
                })
                .await
                .expect("Error with HTTP server!");
        });

        //for debug
        let _ = sleep(Duration::from_secs(1)).await;
        let client = Client::new(Curve::new(key.as_bytes(), &curve), remote, target, id).await?;

        Ok(ExecuteEngine {
            client: client,
            join_handle: join_handle,
            shutdown: shutdown,
        })
    }
}
