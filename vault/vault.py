import json
import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from cryptography.hazmat.primitives.asymmetric import ec
from cryptography.hazmat.primitives import serialization

app = FastAPI()

STORAGE_FILE = "keys.json"

def load_storage():
    if not os.path.exists(STORAGE_FILE):
        return {}
    with open(STORAGE_FILE, "r") as f:
        return json.load(f)


def save_storage(data):
    with open(STORAGE_FILE, "w") as f:
        json.dump(data, f, indent=2)


def generate_keypair():
    private_key = ec.generate_private_key(ec.SECP256R1())
    public_key = private_key.public_key()

    private_pem = private_key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption()
    ).decode()

    public_pem = public_key.public_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PublicFormat.SubjectPublicKeyInfo
    ).decode()

    return private_pem, public_pem


class KeyRequest(BaseModel):
    name: str


@app.get("/key/{name}")
def get_key(name: str):
    data = load_storage()

    if name not in data:
        private_pem, public_pem = generate_keypair()

        data[name] = {"private_key": private_pem, "public_key": public_pem}

        save_storage(data)

    return data[name]