use std::env;

use tokio::sync::OnceCell;

use crate::{Db, Tx, connect};

static GLOBAL_CONN: OnceCell<Db> = OnceCell::const_new();

async fn get_conn() -> Db {
    GLOBAL_CONN
        .get_or_init(async || {
            let url = env::var("DATABASE_URL").expect("DATABASE_URL not set");

            connect(&url).await.expect("error connecting to database")
        })
        .await
        .clone()
}

// Get a test transaction on the already existing global pool.
pub async fn get_test_tx<'a>() -> Tx<'a> {
    get_conn()
        .await
        .begin()
        .await
        .expect("error beginning test transaction")
}
