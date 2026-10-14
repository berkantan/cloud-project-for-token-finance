# Postmortem: Disk dolması nedeniyle sunucuya erişilemedi

**Tarih:** 13 Ekim 2026
**Etkilenen sistem:** token-finans-proje (GCP Compute Engine, e2-micro, us-central1-a)
**Erişim kaybı süresi:** Yaklaşık 1 saat 50 dakika (18:03 – 19:52, TSİ)

## Özet

Sunucunun 10 GB'lık diski tamamen doldu. Disk dolunca SSH anahtarı sunucuya yazılamadı ve sunucuya erişilemedi. Docker servisi de başlatılamadı. Disk, kurtarma VM'i üzerinden 20 GB'a genişletilerek tekrar ssh bağlantısı sağlandı. Kök neden, multi-stage build kullanılmadan oluşturulan ve gereksiz yere büyük olan Docker imajlarıydı.

## Etki

- SSH ile sunucuya erişilemedi.
- Docker servisi başlatılamadığı için payment API, Prometheus ve Grafana çalışmadı.
- Bu süre boyunca monitoring verisi toplanamadı.

## Zaman çizelgesi (TSİ)

| Saat  | Olay                                                                                                                                   |
| ----- | -------------------------------------------------------------------------------------------------------------------------------------- |
| 18:03 | Sunucu açılışında serial log'da `No space left on device` hataları görülmeye başlandı                                                  |
| 18:13 | SSH ile bağlanılamadığı fark edildi. Serial log incelendi, diskin dolu olduğu tespit edildi                                            |
| 18:17 | SSH troubleshoot aracı yüksek CPU uyarısı verdi. Uyarı, asıl sorun olan disk dolmasını gözden kaçırmaya yol açabilecek bir yan etkiydi |
| 18:25 | Temizlik için eklenen startup script çalışmadı, çünkü disk dolu olduğu için script'in kendisi de diske yazılamadı                      |
| 18:41 | İlk kurtarma denemesi: disk 20 GB'a büyütüldü ama dosya sistemi genişletilmediği için sorun devam etti                                 |
| 19:46 | İkinci kurtarma denemesi: disk kurtarma VM'ine takıldı, `growpart` ve `resize2fs` ile dosya sistemi genişletildi                       |
| 19:52 | SSH erişimi geri geldi                                                                                                                 |
| 19:58 | Kök neden bulundu: Docker imajları toplam 4.2 GB yer kaplıyordu                                                                        |
| 20:18 | Uygulama imajı multi-stage build ile yeniden yazıldı                                                                                   |
| 01:14 | Ayrı bulgu: RAM'de 62 MB boş yer kaldığı ve swap olmadığı görüldü. 2 GB swap eklendi                                                   |
| 01:31 | Ayrı bulgu: Prometheus'un uygulamayı hiç scrape etmediği fark edildi ve düzeltildi                                                     |

## Kök neden (5 Neden)

**1. Neden SSH ile bağlanılamadı?**
SSH ile bağlanmak istediğimizde sunucuya SSH anahtarı dosya olarak yazılır. Disk dolunca SSH anahtarı dosyası yazılamadı, bu yüzden SSH beni reddetti (`Permission denied (publickey)`).

**2. Neden disk doldu?**
Diskin dolmasının en büyük sebebi Docker imajlarıydı. Normalde 10-20 MB olabilecek uygulama imajının içine Go'nun kaynak kodu, compiler'ı vs. de girdi ve boyutu 1.5 GB'a kadar yükseldi. Build için çekilen ve artık kullanılmayan `golang` imajı da 1.3 GB yer kaplıyordu. İmajlar toplamda 4.2 GB'a ulaştı.

**3. Neden imajlar bu kadar yer kaplıyordu?**
İmajın içinde kaynak kod ile birlikte build alınmış olması ve multi-stage image çıkarılmaması. Dockerfile yazılırken sadece çalışmasına odaklanıldı, imaj boyutu kontrol edilmedi.

**4. Neden disk dolarken bu önceden fark edilmedi?**
Herhangi bir sınırlayıcı veya alarm üretici mekanizma kullanılmıyordu. Disk kullanımı için alarm yoktu, container logları ve Prometheus verisi için boyut limiti yoktu.

**5. Neden disk dolunca her şey birden durdu?**
Çünkü işletim sistemi ve Docker aynı diski kullanıyordu. Disk Docker imajları yüzünden o kadar doluydu ki işletim sistemi düzgün çalışır duruma gelemedi. İşletim sistemi çalışamayınca container'lar da dolaylı yoldan çalışamamış oldu.

Docker verisi ayrı bir diskte olsaydı ve o disk dolsaydı, işletim sistemi ayakta kalırdı. Sunucuya normal şekilde bağlanıp temizlik yapabilirdik. Docker'ın bulunduğu diske yazma yapılamazdı, ama kurtarma VM'ine gerek kalmazdı. Yani sorunun etki alanı (blast radius) daha küçük olurdu.

## Tespit

Sorun bir alarmla değil, SSH ile bağlanmaya çalışırken fark edildi. GCP'nin SSH troubleshoot aracı hata göstermedi, sadece yüksek CPU uyarısı verdi. Asıl teşhis serial port loglarından konuldu.

## Çözüm

1. Disk GCP üzerinden 10 GB'tan 20 GB'a büyütüldü.
2. Makine durdurulup diski ayrıldı ve geçici bir kurtarma VM'ine takıldı.
3. Kurtarma VM'inde `growpart` ile bölüm, `resize2fs` ile dosya sistemi genişletildi.
4. Disk asıl makineye geri takıldı ve SSH erişimi geri geldi.
5. Kullanılmayan `golang` imajı silindi.
6. Uygulama Dockerfile'ı multi-stage build ve distroless imaja geçirildi (1.5 GB'tan ~20 MB'a).

## Neler iyi gitti

- Serial port logları sayesinde SSH olmadan da teşhis konulabildi.
- Kurtarma VM yöntemiyle veri kaybı olmadan erişim geri kazanıldı.
- Kök neden sadece semptom çözülerek bırakılmadı. Bulundu ve kalıcı olarak düzeltildi.

## Neler kötü gitti

- Disk dolmasını haber verecek bir alarm yoktu.
- İlk kurtarma denemesinde dosya sistemi genişletme adımı atlandı. Doğrulama yapılmadan bir sonraki adıma geçildiği için yaklaşık bir saat kaybedildi.
- Monitoring kurulu sanılıyordu, ama Prometheus config'i hiç yüklenmemişti ve uygulama izlenmiyordu. Bu, sorun anında monitoring'in zaten işe yaramayacağı anlamına geliyordu.

## Aksiyonlar

| Aksiyon                                                             | Durum      |
| ------------------------------------------------------------------- | ---------- |
| Uygulama imajını multi-stage build ile küçültmek                    | Tamamlandı |
| Kullanılmayan imajları temizlemek                                   | Tamamlandı |
| Container loglarına boyut limiti koymak (10 MB × 3 dosya)           | Tamamlandı |
| Prometheus verisine 1 GB saklama limiti koymak                      | Tamamlandı |
| 2 GB swap eklemek                                                   | Tamamlandı |
| Prometheus config'ini düzeltip uygulamayı gerçekten izlemek         | Tamamlandı |
| Build'i VM'den çıkarıp registry (Docker Hub) üzerinden deploy etmek | Tamamlandı |
| Disk ve hata oranı için alarm kurmak                                | Yapılacak  |
| Docker verisini işletim sisteminden ayrı bir diske taşımak          | Yapılacak  |
| Build ve deploy'u CI/CD pipeline ile otomatikleştirmek              | Yapılacak  |
