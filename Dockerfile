# GO dilinin 1.27.1 sürümünü kur.
FROM golang:1.27.1

# /app olarak çalışma dizinini belirle.
WORKDIR /app

# Gerekli dosyaları /app dizinine kopyala.
COPY go.mod main.go go.sum /app/

# Harici dependency varsa ekle.
RUN go mod download

# Kodu derle
RUN go build -o payment-api main.go

# Yayına alacağımız port numarası.
EXPOSE 9093

# Konteyner ayağa kalktığında çalıştır.
CMD [ "./payment-api" ]