use actix_session::Session;
use actix_web::{get, post, web, HttpResponse, Responder};
use deadpool_postgres::{Client, Pool};
use serde::{Deserialize, Serialize};
use serde_json::json;

use crate::service::{
    check_credentials, get_user_paths, insert_user, read_file, reindex_user_files,
};

#[derive(Deserialize)]
pub struct AuthCredentials {
    username: String,
    password: String,
}

#[derive(Deserialize)]
pub struct ReadFileQuery {
    path: String,
    token: String,
}

#[derive(Deserialize, Serialize)]
pub struct UserData {
    username: String,
    id: i32,
}

fn error<T: std::fmt::Display>(e: T) -> HttpResponse {
    HttpResponse::PreconditionFailed().json(json!({ "error": format!("{}", e) }))
}

fn error_unauthenticated<T: std::fmt::Display>(e: T) -> HttpResponse {
    HttpResponse::Unauthorized().json(json!({ "error": format!("{}", e) }))
}

#[post("/api/login")]
pub async fn login(
    req: web::Json<AuthCredentials>,
    session: Session,
    db_pool: web::Data<Pool>,
) -> impl Responder {
    let creds = req.into_inner();

    let client: Client = match db_pool.get().await {
        Ok(cli) => cli,
        Err(e) => return error(e),
    };

    match check_credentials(&client, &creds.username, &creds.password).await {
        Ok(user_id) => {
            let udata = UserData {
                username: creds.username,
                id: user_id,
            };

            match session.insert("user", &udata) {
                Ok(_) => HttpResponse::Ok().json(udata),
                Err(e) => error(e),
            }
        }
        Err(e) => error(e),
    }
}

#[post("/api/register")]
pub async fn register(req: web::Json<AuthCredentials>, db_pool: web::Data<Pool>) -> impl Responder {
    let creds = req.into_inner();

    let client: Client = match db_pool.get().await {
        Ok(cli) => cli,
        Err(e) => return error(e),
    };

    match insert_user(&client, &creds.username, &creds.password).await {
        Ok(_) => HttpResponse::Ok().json("registered"),
        Err(err) => error(err),
    }
}

#[get("/api/file")]
pub async fn get_file(
    read_query: web::Query<ReadFileQuery>,
    session: Session,
    db_pool: web::Data<Pool>,
) -> impl Responder {
    let user = match session.get("user") {
        Err(e) => return error_unauthenticated(e),
        Ok(value) => value,
    };
    if user.is_none() {
        return error_unauthenticated("No user saved in session");
    }

    let user: UserData = user.unwrap();

    let client: Client = match db_pool.get().await {
        Ok(cli) => cli,
        Err(e) => return error(e),
    };

    match read_file(&client, &user.username, &read_query.path, &read_query.token).await {
        Ok(content) => HttpResponse::Ok().body(content),
        Err(e) => error(e),
    }
}

#[get("/api/files")]
pub async fn get_indexed_files(session: Session, db_pool: web::Data<Pool>) -> impl Responder {
    let user = match session.get("user") {
        Err(e) => return error_unauthenticated(e),
        Ok(value) => value,
    };
    if user.is_none() {
        return error_unauthenticated("No user saved in session");
    }

    let user: UserData = user.unwrap();

    let client: Client = match db_pool.get().await {
        Ok(cli) => cli,
        Err(e) => return error(e),
    };

    match get_user_paths(&client, &user.username).await {
        Ok(paths) => HttpResponse::Ok().json(paths),
        Err(e) => error(e),
    }
}

#[post("/api/reindex")]
pub async fn reindex(session: Session, db_pool: web::Data<Pool>) -> impl Responder {
    let user = match session.get("user") {
        Err(e) => return error_unauthenticated(e),
        Ok(value) => value,
    };
    if user.is_none() {
        return error_unauthenticated("No user saved in session");
    }

    let user: UserData = user.unwrap();

    let client: Client = match db_pool.get().await {
        Ok(cli) => cli,
        Err(e) => return error(e),
    };

    match reindex_user_files(&client, &user.username).await {
        Ok(_) => HttpResponse::Ok().json("reindexed"),
        Err(e) => error(e),
    }
}
