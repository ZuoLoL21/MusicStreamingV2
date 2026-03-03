"""
MinIO S3 storage operations for file uploads
"""

import os
import mimetypes
from pathlib import Path
import boto3
from botocore.client import Config
from botocore.exceptions import ClientError


def connect_minio():
    """Connect to MinIO S3 service"""

    endpoint = os.environ.get('MINIO_ENDPOINT', 'localhost:9000')
    access_key = os.environ.get('MINIO_ACCESS_KEY', 'minioadmin')
    secret_key = os.environ.get('MINIO_SECRET_KEY', 'minioadmin')
    use_ssl = os.environ.get('MINIO_USE_SSL', 'false').lower() == 'true'

    # Determine endpoint URL
    protocol = 'https' if use_ssl else 'http'
    endpoint_url = f"{protocol}://{endpoint}"

    # Create S3 client
    s3_client = boto3.client(
        's3',
        endpoint_url=endpoint_url,
        aws_access_key_id=access_key,
        aws_secret_access_key=secret_key,
        config=Config(signature_version='s3v4'),
        region_name='us-east-1'  # MinIO doesn't care about region
    )

    return s3_client


def ensure_bucket_exists(s3_client, bucket_name):
    """Ensure S3 bucket exists, create if not"""
    try:
        s3_client.head_bucket(Bucket=bucket_name)
    except ClientError as e:
        error_code = e.response['Error']['Code']
        if error_code == '404':
            # Bucket doesn't exist, create it
            s3_client.create_bucket(Bucket=bucket_name)
            print(f"Created bucket: {bucket_name}")
        else:
            raise


def upload_audio(s3_client, bucket_name, music_uuid, mp3_path):
    """
    Upload MP3 file to MinIO
    Returns S3 URL path
    """
    # S3 key format: audio/{uuid}.mp3
    s3_key = f"audio/{music_uuid}.mp3"

    # Determine content type
    content_type = 'audio/mpeg'

    # Upload file
    try:
        with open(mp3_path, 'rb') as f:
            s3_client.put_object(
                Bucket=bucket_name,
                Key=s3_key,
                Body=f,
                ContentType=content_type
            )

        # Return the S3 path (not full URL, as app will construct URL)
        return s3_key

    except Exception as e:
        raise Exception(f"Failed to upload audio {mp3_path}: {e}")


def upload_image(s3_client, bucket_name, image_type, entity_uuid, image_path):
    """
    Upload image file to MinIO

    Args:
        image_type: 'artist', 'album', or 'music'
        entity_uuid: UUID of the entity
        image_path: Path to image file

    Returns:
        S3 URL path
    """
    # Determine file extension
    suffix = Path(image_path).suffix.lower()
    if not suffix:
        suffix = '.jpg'  # Default to jpg

    # S3 key format: {type}_images/{uuid}{ext}
    s3_key = f"{image_type}_images/{entity_uuid}{suffix}"

    # Determine content type
    content_type, _ = mimetypes.guess_type(image_path)
    if not content_type:
        content_type = 'image/jpeg'  # Default

    # Upload file
    try:
        with open(image_path, 'rb') as f:
            s3_client.put_object(
                Bucket=bucket_name,
                Key=s3_key,
                Body=f,
                ContentType=content_type
            )

        # Return the S3 path
        return s3_key

    except FileNotFoundError:
        raise Exception(f"Image file not found: {image_path}")
    except Exception as e:
        raise Exception(f"Failed to upload image {image_path}: {e}")


def upload_artist_image(s3_client, bucket_name, artist_uuid, image_path):
    """Upload artist profile image"""
    return upload_image(s3_client, bucket_name, 'artist', artist_uuid, image_path)


def upload_album_image(s3_client, bucket_name, album_uuid, image_path):
    """Upload album cover image"""
    return upload_image(s3_client, bucket_name, 'album', album_uuid, image_path)


def upload_music_image(s3_client, bucket_name, music_uuid, image_path):
    """Upload music track cover image"""
    return upload_image(s3_client, bucket_name, 'music', music_uuid, image_path)


def file_exists(s3_client, bucket_name, s3_key):
    """Check if file exists in S3"""
    try:
        s3_client.head_object(Bucket=bucket_name, Key=s3_key)
        return True
    except ClientError as e:
        error_code = e.response['Error']['Code']
        if error_code == '404':
            return False
        else:
            raise
