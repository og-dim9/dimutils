
.PHONY: curl-setup clean curl

curl: curl-setup
	@echo "Running curl"
	./../gitaskop run --repo "file://$(PWD)/curl" --triggeronchange --clean --branch "main"

curl-setup:
	@mkdir -pv curl
	@git -C curl init && touch curl/date.txt
	@git -C curl add date.txt
	@git -C curl commit -m "Initial commit"
	@./../gitaskop generate job --interval "* * * * *" --script \
		"sh -c 'date >date.txt && git add * && git commit -m f && git push'" \
		| tee curl/.gitaskop.json
	@git -C curl add .gitaskop.json
	@git -C curl commit -m "Add gitaskop job"

clean:
	@rm -rvf curl


