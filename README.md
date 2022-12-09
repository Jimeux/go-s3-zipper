# go-s3-zipper

## Bucket Setup

1. Set a profile in `~/.aws/credentials`
2. Create buckets
    ```
    cd terraform
    terraform init
    terraform apply
    cd -
    ```
3. Set bucket names in `.env` (terraform outputs)
4. Run `direnv allow`
5. Run `make sync-images`

## Run Code

`make run`

## Bucket Cleanup
```
make clear-buckets
cd terraform
terraform destroy
```
