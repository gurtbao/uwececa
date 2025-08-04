use crate::{Reply, templates::render};

pub async fn index() -> Reply {
    render("main/home.html", Some("Home".to_owned()), ())
}
