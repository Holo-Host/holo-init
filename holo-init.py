import json
import requests
import retry
import subprocess
import sys

def holochain_keygen(path):
    return subprocess.run(['hc', 'keygen', '-np', path, '-q'], capture_output=True) \
        .stdout \
        .split(b'\n')[0] \
        .decode('utf-8')

HOLO_INIT_KEY = "wbfGXvzmLk83bUmR"

def zato_request(endpoint, payload):
    return requests.post('http://proxy.holohost.net/zato' + endpoint,
            headers={'Holo-Init': HOLO_INIT_KEY},
            json=payload).json()

def zato_setup_dns(public_key):
    return zato_request('/holo-init-cloudflare-dns-create', {'pubkey': public_key})

def zato_setup_zerotier(zerotier_address):
    return zato_request('/holo-zt-auth', {'member_id': zerotier_address})

def zato_setup_zerotier_address():
    return zato_setup_zerotier(zerotier_address())

def zato_setup_proxy(public_key, ipv4):
    return zato_request('/holo-init-proxy-service-create', {
        'name': public_key + '.holohost.net',
        'protocol': 'http',
        'host': ipv4,
        'port': 48080
    })

def zato_setup_proxy_route(public_key, proxy_id):
    return zato_request('/holo-init-proxy-route-create', {
        'name': public_key + '.holohost.net',
        'protocols': ['http', 'https'],
        'hosts': ['*.' + public_key.lower() + '.holohost.net'],
        'service': proxy_id
    })

def zerotier_run(args):
    process = subprocess.run(['zerotier-cli', '-j'] + args, capture_output=True)
    return json.loads(process.stdout)

def zerotier_info():
    return zerotier_run(['info'])

def zerotier_address():
    return zerotier_info()['address']

@retry.retry(IndexError, tries=10, delay=2, backoff=2)
def zerotier_ipv4():
    return zato_setup_zerotier_address()['config']['ipAssignments'][0]

def main(private_key_path):
    ipv4 = zerotier_ipv4()
    public_key = holochain_keygen(private_key_path)
    with open(private_key_path + '.pub', 'w') as f:
        print(public_key, file=f)

    zato_setup_dns(public_key)
    res = zato_setup_proxy(public_key, ipv4)
    print(zato_setup_proxy_route(public_key, res['id'])['name'])

if __name__ == "__main__":
    main(sys.argv[1])
