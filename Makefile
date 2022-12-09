.PHONY: sync-images
sync-images:
	aws s3 sync --acl public-read --region ap-northeast-1 ./images s3://$(IMAGE_BUCKET) --

.PHONY: delete-images
clear-buckets:
	aws s3 rm s3://$(IMAGE_BUCKET) --recursive
	aws s3 rm s3://$(UPLOAD_BUCKET) --recursive

.PHONY: run
run:
	@go run main.go
