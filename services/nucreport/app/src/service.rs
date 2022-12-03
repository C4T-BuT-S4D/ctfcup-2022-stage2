use std::path::Path;

use anyhow::{bail, ensure, Context, Result};
use deadpool_postgres::Client;
use rand::Rng;
use regex::Regex;
use serde::{Deserialize, Serialize};
use tokio::process::Command;
use tokio::sync::Mutex;
use walkdir::{DirEntry, WalkDir};

static CREATE_MUTEX: Mutex<()> = Mutex::const_new(());

#[derive(Deserialize, Serialize)]
pub struct PathInfo {
    path: String,
    token: String,
}

// 50kb
const MAX_READ_BYTES: u64 = 50 << 10;
const ADD_USER_SCRIPT: &str = "/usr/sbin/addnewuser.sh";
const READ_FILE_SCRIPT: &str = "/usr/sbin/readfile.sh";

pub async fn check_credentials(cli: &Client, user: &str, password: &str) -> Result<i32> {
    let stmt = cli
        .prepare("SELECT id FROM users WHERE username = $1 AND password = $2")
        .await?;

    let res = cli.query(&stmt, &[&user, &password]).await?;
    let row = res.first().context("Invalid credentials")?;
    let user_id: i32 = row.get(0);
    Ok(user_id)
}

pub async fn insert_user(cli: &Client, user: &String, password: &String) -> Result<()> {
    let u_regex = Regex::new(r"(?m)[^a-z\d]+").unwrap();
    let p_regex = Regex::new(r"(?m)[^a-zA-Z\d]+").unwrap();

    ensure!(!u_regex.is_match(user), "Invalid username");
    ensure!(!p_regex.is_match(password), "Invalid password");
    ensure!(user.len() > 5, "Invalid username length");
    ensure!(password.len() > 5, "Invalid username length");

    let stmt = cli
        .prepare("SELECT * FROM users WHERE username = $1")
        .await?;

    let res = cli.query(&stmt, &[user]).await?;
    ensure!(res.first().is_none(), "User already exists");

    cli.execute(
        "INSERT INTO users (username, password) VALUES ($1, $2)",
        &[user, password],
    )
    .await?;

    Ok(())
}

pub async fn create_unix_user(user: &String, password: &String) -> Result<()> {
    let _lock = CREATE_MUTEX.lock().await;

    let res = Command::new("sudo")
        .arg(ADD_USER_SCRIPT)
        .arg(user)
        .arg(password)
        .output()
        .await?;

    ensure!(res.status.success(), "Failed to create system user");
    Ok(())
}

fn home_dir(u: &str) -> String {
    format!("/users/{u}/")
}

pub async fn read_file(cli: &Client, user: &str, path: &str, token: &str) -> Result<String> {
    ensure!(!path.is_empty(), "Empty file path");

    let normalized_path = Path::new(&path).canonicalize()?;
    let normalized = normalized_path.to_str().context("Invalid file path")?;

    if !token.is_empty() {
        return read_by_file_token(cli, path, token).await;
    }

    let user_dir = home_dir(user);
    if normalized.starts_with(&user_dir) {
        return safe_read_file(path).await;
    }

    bail!("Unauthorized")
}

pub async fn reindex_user_files(cli: &Client, user: &str) -> Result<()> {
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

pub async fn get_user_paths(cli: &Client, user: &str) -> Result<Vec<PathInfo>> {
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
        .map_or(false, |s| s.starts_with('.'))
}

fn index_dir(user: &str) -> Vec<String> {
    let mut vec: Vec<String> = Vec::new();

    for entry in WalkDir::new(home_dir(user))
        .max_open(5)
        .max_depth(3)
        .into_iter()
        .filter_entry(|e| !is_hidden(e))
        .filter_map(Result::ok)
        .filter(|e| !e.file_type().is_dir())
    {
        let Some(path) = entry.path().to_str() else { continue };
        vec.push(String::from(path));
        if vec.len() > 50 {
            return vec;
        };
    }
    vec
}

async fn read_by_file_token(cli: &Client, path: &str, token: &str) -> Result<String> {
    cli.query_one(
        "SELECT * FROM file_paths WHERE filepath = $1 AND token = $2",
        &[&path, &token],
    )
    .await?;

    safe_read_file(path).await
}

async fn safe_read_file(path: &str) -> Result<String> {
    let res = Command::new("sudo")
        .arg(READ_FILE_SCRIPT)
        .arg(path)
        .arg(MAX_READ_BYTES.to_string())
        .output()
        .await?;

    String::from_utf8(res.stdout).context("failed to decode utf-8")
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
