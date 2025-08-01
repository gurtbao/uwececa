use std::io::Write;

use minijinja::{Environment, path_loader};
use minijinja_autoreload::{AutoReloader, EnvironmentGuard};
use serde::Serialize;
use uwececa_cfg::is_development;

pub struct Renderer(AutoReloader);

impl Renderer {
    pub fn acquire_env(&self) -> color_eyre::Result<EnvironmentGuard<'_>> {
        Ok(self.0.acquire_env()?)
    }

    pub fn render_write<S: Serialize, W: Write>(
        &self,
        name: &'static str,
        w: W,
        s: S,
    ) -> color_eyre::Result<()> {
        self.acquire_env()?
            .get_template(name)?
            .render_to_write(s, w)?;

        Ok(())
    }

    pub fn render_string<S: Serialize, W: Write>(
        &self,
        name: &'static str,
        s: S,
    ) -> color_eyre::Result<String> {
        Ok(self.acquire_env()?.get_template(name)?.render(s)?)
    }
}

pub fn get_templates<'a>(path: String) -> color_eyre::Result<Renderer> {
    let reloader = AutoReloader::new(move |notifier| {
        let mut env = Environment::new();

        minijinja_contrib::add_to_environment(&mut env);

        env.set_loader(path_loader(&path));
        if is_development() {
            notifier.watch_path(&path, true);
        }

        Ok(env)
    });

    Ok(Renderer(reloader))
}

