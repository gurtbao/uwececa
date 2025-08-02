use std::io::Write;

use minijinja::{Environment, path_loader};
use minijinja_autoreload::{AutoReloader, EnvironmentGuard};
use serde::Serialize;
use uwececa_cfg::{git_hash, is_development};

pub struct Renderer(AutoReloader);
pub use minijinja::context;

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

    pub fn render_string<S: Serialize>(
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

        env.add_global("git_commit_hash", git_hash());

        env.set_loader(path_loader(&path));
        if is_development() {
            notifier.watch_path(&path, true);
        }

        Ok(env)
    });

    Ok(Renderer(reloader))
}

#[cfg(test)]
mod test {
    use minijinja::context;

    use crate::get_templates;

    #[test]
    fn test_templates() {
        let templ = get_templates("./test".into()).unwrap();

        let rendered = templ
            .render_string("test.html", context!(data => "Hallo"))
            .unwrap();

        assert_eq!(rendered, "<h1>Hallo</h1>");
    }
}
