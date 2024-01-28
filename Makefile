.PHONY: docker-builder vlight vic-gateway

docker-builder:
	docker build -t armbuilder docker-builder/.

all: vlight vic-gateway

go_deps:
	echo `go version` && cd `pwd` && go mod download

vic-custom: go_deps
	docker container run  \
	-v $(GOPATH)/pkg/mod:/go/pkg/mod \
	-v "`pwd`":/go/src/digital-dream-labs/vector-cloud \
	-w /go/src/digital-dream-labs/vector-cloud \
	-v /tmp:/tmp \
	--user $(UID):$(GID) \
	armbuilder \
	go build  \
	-tags vicos \
	--trimpath \
	-ldflags '-w -s -linkmode internal -extldflags "-static" -r /anki/lib' \
	-o build/vic-custom \
	*.go


vic-gateway: go_deps
	docker container run \
	-v "$(PWD)":/go/src/digital-dream-labs/vector-cloud \
	-v $(GOPATH)/pkg/mod:/go/pkg/mod \
	-w /go/src/digital-dream-labs/vector-cloud \
	--user $(UID):$(GID) \
	armbuilder \
	go build  \
	-tags nolibopusfile,vicos \
	--trimpath \
	-ldflags '-w -s -linkmode internal -extldflags "-static" -r /anki/lib' \
	-o build/vic-gateway \
	gateway/*.go

	docker container run \
	-v "$(PWD)":/go/src/digital-dream-labs/vector-cloud \
	-v $(GOPATH)/pkg/mod:/go/pkg/mod \
	-w /go/src/digital-dream-labs/vector-cloud \
	--user $(UID):$(GID) \
	armbuilder \
	upx build/vic-gateway
