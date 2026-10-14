# GO dilinin 1.27.1 sürümünü kur.
FROM golang:1.27.1 AS builder

# /app olarak çalışma dizinini belirle.
WORKDIR /app

# Gerekli dosyaları /app dizinine kopyala.
COPY go.mod go.sum ./

# Harici dependency varsa ekle.
RUN go mod download

COPY main.go ./

# Kodu derler ama binary libc ile bağlanmaz. Bunu CGO=0 sağlar. CGO'yu 0 ayarladım çünkü
# distroless'ta libc bile yok. no such file or directory verir 0 ayarlanmazsa.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o payment-api main.go

# 2. FROM komutu ile yeni bir docker imajı başlar, ilk aşamadaki her şey geride kalır. 
# Static distroless varyantı harici kütüphaneye ihtiyaç duymayan binaryler için var.
FROM gcr.io/distroless/static-debian12

# payment-api dosyasını ilk aşamanın file systeminden alır.
COPY --from=builder /app/payment-api /payment-api

EXPOSE 9093

USER nonroot:nonroot

ENTRYPOINT ["/payment-api"]
