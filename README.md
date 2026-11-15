# Nguyễn Quốc Nghi — Go portfolio

Website portfolio độc lập bằng **Go `net/http` + `html/template` + PostgreSQL**, kèm ảnh và CV hiện tại. Giao diện HTML/CSS/JavaScript nhẹ, có dark mode và responsive; không dùng React để giả lập backend Go.

**Trạng thái:** source đã build và chạy kiểm tra local. Chưa deploy lên Internet, chưa tạo tài nguyên cloud. Xem `VERIFICATION.md` để biết chính xác phần đã/chưa kiểm tra; xem `DEPLOY.md` để triển khai.

## Tính năng

- Trang giới thiệu, ảnh chân dung, tải CV PDF, kinh nghiệm, năm dự án và ba khóa học chọn lọc.
- Phân biệt dự án production với Go cá nhân; không tự nâng thâm niên hoặc thêm MLOps.
- Lọc dự án, xem phần đóng góp, dark/light theme lưu trên trình duyệt; nội dung chính vẫn đọc được nếu tắt JavaScript.
- API profile/projects, liveness/readiness và nút kiểm tra phản hồi backend thật. Thời gian một request không được gọi là benchmark.
- Contact form có validation, double-submit CSRF cookie + kiểm tra Origin khi có, honeypot, body limit, quota đơn giản và lưu vào PostgreSQL.
- Go tests, race tests, integration test tùy chọn, GitHub Actions CI, Dockerfile và Compose.

Không có CMS/admin hoặc chức năng gửi email tự động. Chủ website đọc tin qua database; khách có thể dùng link `mailto:` để gửi email trực tiếp. Không tích hợp demo AI gateway khi chưa có source và cấu hình gateway; thẻ dự án chỉ mô tả kinh nghiệm trong CV.

## Cấu trúc

```text
main.go                      cấu hình, khởi động, shutdown và retention
content.go                   kiểu dữ liệu, embedded profile và assets
server.go                    HTTP routes, form, CSRF, security và rate limiter
store.go                     Store interface, memory/PostgreSQL implementations
server_test.go               kiểm tra HTTP, validation, security và concurrency
store_integration_test.go    PostgreSQL integration trong schema tạm
web/index.html               giao diện server-rendered
web/assets/                  CSS, JS, favicon, ảnh và CV
data/profile.json            nội dung public; chỉnh tại đây rồi build/deploy lại
.github/workflows/ci.yml     vet, race/integration tests, build và Docker build
Dockerfile / compose.yaml    container và database local
render.yaml                  mẫu Blueprint có thể tạo tài nguyên trả phí
```

## Chạy nhanh — Go

Cài Go **1.27.x** (module yêu cầu tối thiểu 1.25; CI/Docker dùng 1.27). Tại thư mục dự án:

```bash
go mod download
go run .
```

Mở `http://localhost:8080`. Không có `DATABASE_URL` và không bật production thì ứng dụng chạy **memory mode có cảnh báo rõ**, chỉ để phát triển; tin nhắn mất khi dừng tiến trình.

Binary không tự đọc `.env`; export biến môi trường trước `go run`, hoặc dùng Compose bên dưới. Không dùng bộ nhớ như database production và không tự chuyển sang bộ nhớ khi PostgreSQL lỗi.

### Với PostgreSQL sẵn có

```bash
export DATABASE_URL='postgres://USER:PASSWORD@HOST:5432/DB?sslmode=require'
export PUBLIC_URL='http://localhost:8080'
go run .
```

Dùng `sslmode=disable` chỉ khi database local hoặc môi trường tin cậy phù hợp; với cloud, theo yêu cầu TLS của nhà cung cấp. Mật khẩu trong URL phải được URL-encode nếu có ký tự đặc biệt. Khi production, đặt `APP_ENV=production` và origin HTTPS thật.

### Với Docker Compose

1. Sao chép `.env.example` thành `.env`.
2. Thay `POSTGRES_PASSWORD` bằng mật khẩu dài, riêng biệt, dùng ký tự URL-safe. Không commit `.env`.
3. Chạy:

```bash
docker compose up --build -d
docker compose logs -f web
```

Web chỉ bind `127.0.0.1:8080` trên máy host; database không mở cổng ra ngoài. Compose này dành cho local, chưa phải cấu hình public có TLS.

## API

| Route | Method | Nội dung |
|---|---|---|
| `/` | GET | Portfolio server-rendered và form token |
| `/api/projects` | GET | Dự án từ Store, PostgreSQL ở production |
| `/api/profile` | GET | Thông tin công khai, không chứa contact submissions |
| `/healthz` | GET | Liveness; không kiểm tra database |
| `/readyz` | GET | Readiness + ping database; 503 nếu không sẵn sàng |
| `/contact` | POST | URL-encoded form; CSRF cookie/token, validation và storage |
| `/assets/resume.pdf` | GET | Bản CV public có thông tin liên hệ |

## Kiểm thử

```bash
go vet ./...
go test -race -count=1 ./...
go build -mod=readonly -trimpath -o portfolio .
```

PostgreSQL integration mặc định **skip** nếu không có `TEST_DATABASE_URL`. Để chạy, cung cấp **database kiểm thử riêng**, tài khoản có quyền tạo/xóa schema:

```bash
export TEST_DATABASE_URL='postgres://USER:PASSWORD@localhost:5432/portfolio_test?sslmode=disable'
go test -race -count=1 -v ./...
```

Test tạo schema ngẫu nhiên rồi xóa; không chạy trên database khách hàng/production. Workflow CI chuẩn bị PostgreSQL riêng, nhưng workflow chưa chạy trên GitHub cho tới khi bạn đẩy repo lên.

## Cấu hình

| Biến | Mặc định | Ý nghĩa |
|---|---|---|
| `PORT` | `8080` | Cổng HTTP, 1–65535 |
| `APP_ENV` | chế độ development | `production` bắt buộc PostgreSQL + HTTPS public origin |
| `PUBLIC_URL` | `http://localhost:PORT` ở development | Origin chính xác; không có path/query, dùng để kiểm tra Origin và canonical |
| `RENDER_EXTERNAL_URL` | không có | Dùng nếu `PUBLIC_URL` trống trên Render |
| `DATABASE_URL` | không có | Có URL thì phải kết nối thành công; không fallback khi lỗi |
| `POSTGRES_PASSWORD` | không có | Chỉ cho Compose; thay trước khi chạy |
| `TEST_DATABASE_URL` | không có | Chỉ cho integration tests; không phải database production |

## Nội dung và dữ liệu

- Chỉnh `data/profile.json` để thay profile, dự án và training; build lại vì nội dung được embed vào binary.
- Project JSON được đồng bộ vào bảng `portfolio_projects` mỗi lần khởi động, kể cả xóa project không còn trong file. Không sửa tay bảng project rồi kỳ vọng thay đổi tồn tại sau redeploy.
- Tin nhắn nằm riêng trong `portfolio_contacts`; không có API public liệt kê tin. Job xóa tin hơn 30 ngày chạy lúc startup và mỗi 24 giờ khi tiến trình hoạt động. Tin có thể tồn tại lâu hơn nếu server ngừng hoạt động; backups theo chính sách nhà cung cấp.
- `web/assets/resume.pdf` chứa số điện thoại/email từ CV; chỉ deploy nếu bạn đồng ý công khai. Thay ảnh tại `web/assets/portrait.jpg` và CV tại đường dẫn trên, rồi build lại.
- `nghi.dev` trên logo chỉ là nhãn thiết kế, **không phải domain đã đăng ký hoặc deploy**. Đổi nhãn tùy ý; phải mua/gắn domain riêng nếu muốn dùng địa chỉ đó.

## Giới hạn và các quyết định có chủ ý

- **Quota liên hệ:** 5 lần/10 phút theo IP kết nối, lưu trong RAM của từng tiến trình. Có giới hạn 4.096 client entries; reset khi restart. Không phải distributed rate limiting. Không tin `X-Forwarded-For` tùy ý; sau proxy, nhiều khách có thể chia quota. Khi public nên bổ sung rate limit ở edge hoặc một cơ chế Redis/shared-store được cấu hình đúng proxy.
- **Form:** lưu DB, không gửi email; không có chống spam nâng cao, retry queue, deduplication hay chống mọi loại bot. Cần hoàn thiện thêm nếu lượng truy cập lớn.
- **Schema:** khởi tạo schema additive trong transaction + advisory lock. Với thay đổi schema tiếp theo, dùng migrations có version và chiến lược rollout/rollback rõ; không đổi schema destructive trên production theo cách này.
- **Triển khai:** Go app phục vụ HTTP sau reverse proxy HTTPS, không tự cấp certificate. Production cookies có Secure/HttpOnly/SameSite=Strict. Origin phải đúng domain chính.
- **Phạm vi:** là portfolio nhỏ, không dùng microservices/k8s/Redis chỉ để thêm keyword; không có tuyên bố đã scale hoặc benchmark tải lớn.
- **Bảo mật:** có biện pháp cơ bản và tests, không phải chứng nhận audit bảo mật. Xem lại dependency advisories, dữ liệu công khai, TLS, quyền database và backup trước khi vận hành thật.

Mã giao diện/Go được viết riêng cho hồ sơ này, không sao chép source mẫu Nikhil. Thư viện bên ngoài tuân theo giấy phép riêng; ảnh, CV và nội dung cá nhân không được coi là dữ liệu mẫu để tái sử dụng.
