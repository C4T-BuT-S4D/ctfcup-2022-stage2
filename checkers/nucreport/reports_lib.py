import io
import tempfile

import re
import requests
import paramiko
from checklib import *

WEB_PORT = 8080
SSH_PORT = 4222


class CheckMachine:
    @property
    def url(self):
        return f'http://{self.c.host}:{self.web_port}'

    def __init__(self, checker: BaseChecker):
        self.c = checker
        self.web_port = WEB_PORT
        self.ssh_port = SSH_PORT

    def register(self, session: requests.Session, username: str, password: str):
        resp = session.post(self.url + '/register', json={'username': username, 'password': password})
        self.c.assert_eq(resp.status_code, 200, 'Failed to register new user: {}, {}'.format(username, password))

    def login(self, session: requests.Session, username: str, password: str, status: Status):
        resp = session.post(self.url + '/login', json={'username': username, 'password': password})
        self.c.assert_eq(resp.status_code, 200, 'Failed to login using web')

    def reindex_files(self, session: requests.Session, status=status.Status.MUMBLE):
        resp = session.post(self.url + '/reindex')
        self.c.assert_eq(resp.status_code, 200, 'Failed to reindex user directory files', status=status)

    def list_user_files_web(self, session: requests.Session):
        resp = session.get(self.url + '/files')
        self.c.assert_eq(resp.status_code, 200, 'Failed to get user indexed files', status=status.Status.MUMBLE)
        return resp.json()

    def get_file_web(self, session: requests.Session, path, token, status=status.Status.MUMBLE):
        resp = session.get(self.url + '/file', params={'path': path, 'token': token})
        self.c.assert_eq(resp.status_code, 200, "Failed to get file from web app", status=status)
        return resp.content

    def put_file(self, username, password, file_name, content):
        cli = paramiko.SSHClient()
        cli.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        try:
            cli.connect(self.c.host, port=self.ssh_port, username=username, password=password)
            sftp_cli = cli.open_sftp()
            sftp_cli.putfo(io.BytesIO(content), file_name)
        except (paramiko.ssh_exception.SSHException, paramiko.ssh_exception.NoValidConnectionsError, IOError) as e:
            self.c.cquit(status.Status.MUMBLE, "Failed to upload file using SFTP", str(e))

    def get_file(self, username, password, file_name):
        cli = paramiko.SSHClient()
        cli.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        try:
            cli.connect(self.c.host, port=self.ssh_port, username=username, password=password)
            sftp_cli = cli.open_sftp()
            with tempfile.TemporaryFile() as f:
                sftp_cli.getfo(file_name, f)
                f.seek(0)
                return f.read()
        except (paramiko.ssh_exception.SSHException, paramiko.ssh_exception.NoValidConnectionsError, IOError):
            return b''