use axum::response::{Html, IntoResponse};
use serde::Serialize;

use crate::{Reply, global_state::global_state};

pub fn render<W: Serialize>(name: &'static str, title: Option<String>, page: W) -> Reply {
    RenderContext::new(title, page).render_reply(name)
}

#[derive(Serialize)]
pub struct RenderContext<W: Serialize> {
    asset_location: &'static str,
    title: Option<String>,
    page: W,
}

impl<W: Serialize> RenderContext<W> {
    pub fn new(title: Option<String>, page: W) -> Self {
        let gs = global_state();

        Self {
            asset_location: &gs.config.assets_location,
            title,
            page,
        }
    }

    pub fn render(&self, name: &'static str) -> color_eyre::Result<String> {
        global_state().templates.render_string(name, self)
    }

    pub fn render_reply(&self, name: &'static str) -> Reply {
        Ok(Html(self.render(name)?).into_response())
    }
}
