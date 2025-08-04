use std::sync::Arc;

use axum::{
    Router,
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::get,
};
use global_state::{GlobalState, set_global_state};
use serde::Serialize;
use templates::render;
use uwececa_cfg::{Config, is_development};
use uwececa_templ::{Renderer, context};
use views::index;

mod global_state;
mod templates;
mod views;

pub fn routes(templ: Arc<Renderer>, config: Config) -> Router {
    let global_state = GlobalState {
        templates: templ,
        config: config.clone(),
    };
    set_global_state(global_state);

    Router::new().route("/", get(index))
}

type Reply = Result<Response, Error>;

trait IntoReply {
    fn into_reply(self) -> Reply;
}

impl<T: IntoResponse> IntoReply for T {
    fn into_reply(self) -> Reply {
        Ok(self.into_response())
    }
}

#[derive(Debug)]
enum Error {
    Internal {
        msg: String,
    },
    WithStatus {
        code: StatusCode,
        msg: &'static str,
        template: Option<&'static str>,
    },
}

impl Error {
    fn from_report(e: color_eyre::Report) -> Self {
        let bt = e.to_string();

        tracing::error!("handler error: {}", bt);

        let mut msg = "Internal Server Error.".into();
        if is_development() {
            msg = format!("{}: {}", msg, bt);
        }

        Self::Internal { msg }
    }

    pub fn new(code: StatusCode, msg: &'static str) -> Self {
        Self::WithStatus {
            code,
            msg,
            template: None,
        }
    }

    pub fn with_template(code: StatusCode, msg: &'static str, template: &'static str) -> Self {
        Self::WithStatus {
            code,
            msg,
            template: Some(template),
        }
    }
}

impl IntoResponse for Error {
    fn into_response(self) -> axum::response::Response {
        match self {
            Self::Internal { msg } => (StatusCode::INTERNAL_SERVER_ERROR, msg).into_response(),
            Self::WithStatus {
                code,
                msg,
                template,
            } => {
                let t = template.unwrap_or("main/error.html");

                render(t, Some(format!("Error: {}", code)), context! {msg=>msg}).into_response()
            }
        }
    }
}

impl From<color_eyre::Report> for Error {
    fn from(value: color_eyre::Report) -> Self {
        Self::from_report(value)
    }
}
