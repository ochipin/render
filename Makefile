all:
	touch testfile.text
	chmod 100 testfile.text
	go test -v -cover -coverprofile cover.out
	go tool cover -func=cover.out
	chmod 664 testfile.text
	rm testfile.text

html:
	touch testfile.text
	chmod 100 testfile.text
	go test -cover -coverprofile cover.out
	go tool cover -html=cover.out
	chmod 664 testfile.text
	rm testfile.text

clean:
	rm cover.out