use std::fmt::Display;

use actix_session::Session;
use actix_web::{get, post, web, HttpResponse, HttpResponseBuilder, Responder};
use deadpool_postgres::{Client, Pool};
use serde::{Deserialize, Serialize};
use serde_json::json;

use crate::service::{
    check_credentials, create_unix_user, get_user_paths, insert_user, read_file, reindex_user_files,
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

#[derive(Debug)]
pub struct AppError {
    status: actix_web::http::StatusCode,
    message: String,
}

impl Display for AppError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.message)
    }
}

fn error<T: Display>(e: T) -> AppError {
    AppError {
        status: actix_web::http::StatusCode::PRECONDITION_FAILED,
        message: e.to_string(),
    }
}

fn error_unauthenticated<T: Display>(e: T) -> AppError {
    AppError {
        status: actix_web::http::StatusCode::UNAUTHORIZED,
        message: e.to_string(),
    }
}

impl actix_web::ResponseError for AppError {
    fn status_code(&self) -> actix_web::http::StatusCode {
        self.status
    }

    fn error_response(&self) -> HttpResponse {
        HttpResponseBuilder::new(self.status_code()).json(json!({"error": self.message}))
    }
}

#[post("/api/login")]
pub async fn login(
    req: web::Json<AuthCredentials>,
    session: Session,
    db_pool: web::Data<Pool>,
) -> Result<impl Responder, AppError> {
    let creds = req.into_inner();

    let client: Client = db_pool.get().await.map_err(error)?;

    let user_id = check_credentials(&client, &creds.username, &creds.password)
        .await
        .map_err(error)?;
    let udata = UserData {
        username: creds.username,
        id: user_id,
    };

    session.insert("user", &udata).map_err(error)?;
    Ok(HttpResponse::Ok().json(udata))
}

#[post("/api/register")]
pub async fn register(
    req: web::Json<AuthCredentials>,
    db_pool: web::Data<Pool>,
) -> Result<impl Responder, AppError> {
    let creds = req.into_inner();

    let client: Client = db_pool.get().await.map_err(error)?;

    insert_user(&client, &creds.username, &creds.password)
        .await
        .map_err(error)?;

    create_unix_user(&creds.username, &creds.password)
        .await
        .map_err(error)?;

    Ok(HttpResponse::Ok().json("registered"))
}

#[get("/api/file")]
pub async fn get_file(
    read_query: web::Query<ReadFileQuery>,
    session: Session,
    db_pool: web::Data<Pool>,
) -> Result<impl Responder, AppError> {
    let Some(user) = session.get::<UserData>("user").map_err(error_unauthenticated)? else {
        return Err(error_unauthenticated("No user saved in session"));
    };

    let client: Client = db_pool.get().await.map_err(error)?;

    let content = read_file(&client, &user.username, &read_query.path, &read_query.token)
        .await
        .map_err(error)?;
    Ok(HttpResponse::Ok().body(content))
}

#[get("/api/files")]
pub async fn get_indexed_files(
    session: Session,
    db_pool: web::Data<Pool>,
) -> Result<impl Responder, AppError> {
    let Some(user) = session.get::<UserData>("user").map_err(error_unauthenticated)? else {
        return Err(error_unauthenticated("No user saved in session"));
    };

    let client: Client = db_pool.get().await.map_err(error)?;

    let paths = get_user_paths(&client, &user.username)
        .await
        .map_err(error)?;
    Ok(HttpResponse::Ok().json(paths))
}

#[post("/api/reindex")]
pub async fn reindex(
    session: Session,
    db_pool: web::Data<Pool>,
) -> Result<impl Responder, AppError> {
    let Some(user) = session.get::<UserData>("user").map_err(error)? else {
        return Err(error_unauthenticated("No user saved in session"));
    };

    let client: Client = db_pool.get().await.map_err(error)?;

    reindex_user_files(&client, &user.username)
        .await
        .map_err(error)?;
    Ok(HttpResponse::Ok().json("reindexed"))
}
