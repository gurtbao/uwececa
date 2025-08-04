use std::sync::{Arc, OnceLock};

use uwececa_cfg::Config;
use uwececa_templ::Renderer;

pub struct GlobalState {
    pub config: Config,
    pub templates: Arc<Renderer>,
}

static GLOBAL_STATE: OnceLock<&'static GlobalState> = OnceLock::new();

pub fn global_state() -> &'static GlobalState {
    GLOBAL_STATE
        .get()
        .expect("mainsite global state was not initialized")
}

pub fn set_global_state(state: GlobalState) {
    let state = Box::leak(Box::new(state));

    GLOBAL_STATE
        .set(state)
        .ok()
        .expect("attempted to set mainsite global state multiple times")
}
