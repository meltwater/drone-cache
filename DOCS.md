Use this plugin for caching build artifacts to speed up your build times. This
plugin can create and restore caches of any folders.

## Config

The following parameters are used to configure the plugin:

* **endpoint** - custom endpoint URL (optional, to use a S3 compatible non-Amazon service)
* **access_key** - amazon key (optional)
* **secret_key** - amazon secret key (optional)
* **bucket** - bucket name
* **region** - bucket region (`us-east-1`, `eu-west-1`, etc)
* **encryption** - if provided, use server-side encryption (`AES256`, `aws:kms`, etc)
* **acl** - access to files that are uploaded (`private`, `public-read`, etc)
* **path_style** - whether path style URLs should be used (true for minio, false for aws)
* **mount**   - one or an array of folders to cache
* **rebuild** - boolean flag to trigger a rebuild
* **restore** - boolean flag to trigger a restore

The following secret values can be set to configure the plugin.

* **AWS_ACCESS_KEY_ID** - corresponds to **access_key**
* **AWS_SECRET_ACCESS_KEY** - corresponds to **secret_key**
* **S3_BUCKET** - corresponds to **bucket**
* **S3_REGION** - corresponds to **region**
* **PLUGIN_ENDPOINT** - corresponds to **endpoint**

## Example

The following is a sample configuration in your .drone.yml file:

```yaml
pipeline:
  s3_cache:
    bucket: my-drone-bucket
    image: meltwater/drone-s3-cache
    restore: true
  	mount:
  	  - node_modules

  build:
    image: node:latest
    commands:
      - npm install

  s3_cache:
    bucket: my-drone-bucket
    image: meltwater/drone-s3-cache
    rebuild: true
  	mount:
  	  - node_modules
```
