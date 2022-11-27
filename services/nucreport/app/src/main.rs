use std::sync::Mutex;

use ::config::Config;
use actix_session::{storage::CookieSessionStore, SessionMiddleware};
use actix_web::{cookie, web, App, HttpServer};
use deadpool_postgres::tokio_postgres::NoTls;
use dotenv::dotenv;
use serde::Deserialize;

use crate::routes::{get_file, get_indexed_files, login, register, reindex};

mod routes;
mod service;

const SECRET_KEY: &'static str = include_str!("/tmp/secret_key");

#[derive(Debug, Default, Deserialize)]
pub struct ExampleConfig {
    pub pg: deadpool_postgres::Config,
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv().ok();

    let config_ = Config::builder()
        .add_source(::config::Environment::default())
        .build()
        .unwrap();

    let config: ExampleConfig = config_.try_deserialize().unwrap();

    let pool = config.pg.create_pool(None, NoTls).unwrap();

    HttpServer::new(move || {
        App::new()
            .wrap(
                SessionMiddleware::builder(
                    CookieSessionStore::default(),
                    cookie::Key::from(SECRET_KEY.as_bytes()),
                )
                .cookie_secure(false)
                .build(),
            )
            .app_data(web::Data::new(pool.clone()))
            .service(login)
            .service(register)
            .service(get_file)
            .service(reindex)
            .service(get_indexed_files)
    })
        .workers(10)
    .bind(("0.0.0.0", 8080))?
    .run()
    .await
}
