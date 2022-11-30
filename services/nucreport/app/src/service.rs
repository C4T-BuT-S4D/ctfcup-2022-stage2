use std::error::Error;
use std::path::Path;
use std::process;
use std::sync::Mutex;

use async_process::Command;
use deadpool_postgres::Client;
use rand::Rng;
use regex::Regex;
use serde::{Deserialize, Serialize};
use simple_error::SimpleError;
use walkdir::{DirEntry, WalkDir};

static CREATE_MUTEX: Mutex<i32> = Mutex::new(0);

#[derive(Deserialize, Serialize)]
pub struct PathInfo {
    path: String,
    token: String,
}

// 50kb
const MAX_READ_BYTES: u64 = 50 << 10;
const ADD_USER_SCRIPT: &str = "/usr/sbin/addnewuser.sh";
const READ_FILE_SCRIPT: &str = "/usr/sbin/readfile.sh";

pub async fn check_credentials(
    cli: &Client,
    user: &String,
    password: &String,
) -> Result<i32, Box<dyn Error>> {
    let stmt = cli
        .prepare("SELECT id FROM users WHERE username = $1 AND password = $2")
        .await?;

    let res = cli.query(&stmt, &[user, password]).await?;
    match res.first() {
        None => Err(SimpleError::new("Invalid credentials").into()),
        Some(row) => {
            let user_id: i32 = row.get(0);
            Ok(user_id)
        }
    }
}

pub async fn insert_user(
    cli: &Client,
    user: &String,
    password: &String,
) -> Result<(), Box<dyn Error>> {
    let regex = Regex::new(r"(?m)[^a-zA-Z\d]+").unwrap();

    if regex.is_match(user) {
        return Err(SimpleError::new("Invalid username").into());
    }
    if regex.is_match(password) {
        return Err(SimpleError::new("Invalid password").into());
    }

    let stmt = cli
        .prepare("SELECT * FROM users WHERE username = $1")
        .await?;

    let res = cli.query(&stmt, &[user]).await?;

    if res.first().is_some() {
        return Err(SimpleError::new("User already exists").into());
    }

    cli.execute(
        "INSERT INTO users (username, password) VALUES ($1, $2)",
        &[user, password],
    )
    .await?;

    Ok(())
}

pub fn create_unix_user(user: &String, password: &String) -> Result<(), SimpleError> {
    let _lock = match CREATE_MUTEX.lock() {
        Ok(guard) => guard,
        Err(poisoned) => poisoned.into_inner(),
    };

    process::Command::new("sudo")
        .arg(ADD_USER_SCRIPT)
        .arg(user)
        .arg(password)
        .output()
        .map(|_| ())
        .map_err(|e| SimpleError::new(e.to_string()))
}

fn home_dir(u: &String) -> String {
    format!("/users/{u}/")
}

pub async fn read_file(
    cli: &Client,
    user: &String,
    path: &String,
    token: &String,
) -> Result<String, Box<dyn Error>> {
    if path.is_empty() {
        return Err(SimpleError::new("Empty file path").into());
    }

    let normalized_path = Path::new(&path).canonicalize()?;
    let normalized = normalized_path.to_str();
    if normalized.is_none() {
        return Err(SimpleError::new("Invalid file path").into());
    }
    let normalized = normalized.unwrap();

    if !token.is_empty() {
        return read_by_file_token(cli, path, token).await;
    }

    let user_dir = home_dir(user);
    if normalized.starts_with(&user_dir) {
        return safe_read_file(path).await;
    }

    Err(SimpleError::new("Unauthorized").into())
}

pub async fn reindex_user_files(cli: &Client, user: &String) -> Result<(), Box<dyn Error>> {
    let indexed = index_dir(user);

    for path in indexed {
        let tok = rnd_token();
        cli.execute(
            "INSERT INTO file_paths (filepath, token) VALUES ($1, $2) ON CONFLICT DO NOTHING",
            &[&path, &tok],
        )
        .await?;
    }

    Ok(())
}

pub async fn get_user_paths(cli: &Client, user: &String) -> Result<Vec<PathInfo>, Box<dyn Error>> {
    let stmt = cli
        .prepare("SELECT filepath, token FROM file_paths WHERE filepath LIKE $1")
        .await?;

    let q = home_dir(user) + "%";

    let rows = cli.query(&stmt, &[&q]).await?;

    let mut res: Vec<PathInfo> = Vec::new();
    for row in rows {
        res.push(PathInfo {
            path: row.get(0),
            token: row.get(1),
        });
    }

    Ok(res)
}

fn is_hidden(entry: &DirEntry) -> bool {
    entry
        .file_name()
        .to_str()
        .map(|s| s.starts_with('.'))
        .unwrap_or(false)
}

fn index_dir(user: &String) -> Vec<String> {
    let mut vec: Vec<String> = Vec::new();

    for entry in WalkDir::new(home_dir(user))
        .max_open(5)
        .max_depth(3)
        .into_iter()
        .filter_entry(|e| !is_hidden(e))
        .filter_map(Result::ok)
        .filter(|e| !e.file_type().is_dir())
    {
        let path = entry.path().to_str();
        match path {
            None => {
                continue;
            }
            Some(v) => {
                vec.push(String::from(v));
                if vec.len() > 50 {
                    return vec;
                }
            }
        }
    }
    vec
}

async fn read_by_file_token(
    cli: &Client,
    path: &String,
    token: &String,
) -> Result<String, Box<dyn Error>> {
    cli.query_one(
        "SELECT * FROM file_paths WHERE filepath = $1 AND token = $2",
        &[path, token],
    )
    .await?;

    safe_read_file(path).await
}

async fn safe_read_file(path: &str) -> Result<String, Box<dyn Error>> {
    let res = Command::new("sudo")
        .arg(READ_FILE_SCRIPT)
        .arg(path)
        .arg(MAX_READ_BYTES.to_string())
        .output()
        .await?;

    String::from_utf8(res.stdout)
        .map_err(|_| SimpleError::new("failed to decode utf-8").into())
}

const CHARSET: &[u8] = b"ABCDEFGHIJKLMNOPQRSTUVWXYZ\
                            abcdefghijklmnopqrstuvwxyz\
                            0123456789";
const TOKEN_LEN: usize = 30;

fn rnd_token() -> String {
    let mut rng = rand::thread_rng();
    let token: String = (0..TOKEN_LEN)
        .map(|_| {
            let idx = rng.gen_range(0..CHARSET.len());
            CHARSET[idx] as char
        })
        .collect();
    token
}
