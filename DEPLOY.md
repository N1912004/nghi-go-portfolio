# Triển khai portfolio Go của Nghi

## Trạng thái hiện tại

**Chưa deploy công khai, chưa có URL live, chưa tạo repo hoặc database cloud và chưa phát sinh phí từ thao tác triển khai.** Source đã build/chạy local; PostgreSQL integration và Docker build chưa được chạy trong môi trường hiện tại. Cần kiểm tra hai bước đó trước khi gọi hệ thống production-ready.

Bạn cần một tài khoản GitHub và nơi chạy Go/container + PostgreSQL. Không gửi mật khẩu, token hoặc database connection string vào chat; đăng nhập và đặt secrets trực tiếp trong dashboard của nhà cung cấp.

## Cách đề xuất: GitHub + Render

Đây là một lựa chọn triển khai, không phải yêu cầu của Vulcan Labs. Có thể dùng Cloud Run hoặc VPS thay thế.

### 1. Chuẩn bị repo

1. Giải nén gói source.
2. Tạo repo GitHub của bạn, ví dụ `nghi-go-portfolio`.
3. Đưa **nội dung bên trong thư mục source** vào gốc repo. `go.mod`, `Dockerfile` và `render.yaml` phải ở gốc, không lồng thêm một thư mục nữa.
4. Đảm bảo có cả `.github/workflows/ci.yml`, `.gitignore` và `.env.example`; không đẩy `.env`, mật khẩu hoặc source công ty.
5. Kiểm tra email, ảnh, CV và quyền công khai thông tin dự án trước khi publish. Nhãn logo `nghi.dev` không có nghĩa domain đó thuộc về bạn.

### 2. Chạy CI trước

Trong GitHub → Actions, kiểm tra workflow **Go checks**:

- Formatting và `go vet`.
- Unit/race tests.
- PostgreSQL integration với database tạm của CI.
- Build Go binary và Docker image.

Workflow này được cung cấp dưới dạng cấu hình, **chưa được xác nhận đã chạy trên GitHub**. Chỉ tiếp tục khi bạn xem log và các bước đều pass. Không lấy kết quả test local-memory thay cho integration tests.

### 3. Tạo Blueprint trên Render

1. Tự đăng nhập [Render Dashboard](https://dashboard.render.com/).
2. Chọn tạo Blueprint, kết nối GitHub và chọn repo của bạn.
3. Render đọc `render.yaml` ở gốc repo.
4. **Dừng lại ở màn hình review để kiểm tra giá trước khi bấm tạo.** Mẫu hiện dùng web plan `0.5c-512mb` và PostgreSQL plan `0.1c-256mb` ở Singapore; đây là cấu hình có thể tính phí, không cam kết miễn phí. Giá và khả năng cung cấp thay đổi theo nhà cung cấp/tài khoản.
5. Chỉ phê duyệt nếu bạn đồng ý chi phí, hoặc đổi plan/phương án phù hợp trước khi áp dụng. Không cần gói lớn cho portfolio này.
6. Blueprint liên kết database qua `DATABASE_URL`. Database chặn public inbound IP theo mẫu; web kết nối bằng private network.
7. Khi tạo dịch vụ mới, Render cung cấp URL HTTPS; ứng dụng đọc `RENDER_EXTERNAL_URL` nếu `PUBLIC_URL` chưa được đặt.
8. Kiểm tra build/start logs và health check `/readyz`.

Mẫu tắt auto-deploy (`autoDeployTrigger: 'off'`) để tránh đẩy thay đổi chưa review. Lần tạo đầu tiên vẫn triển khai dịch vụ. Khi CI đã ổn định, có thể đổi sang `checksPass` trong Blueprint nếu bạn muốn deploy sau khi checks đạt.

Tài liệu tham chiếu: [Render Blueprint specification](https://render.com/docs/blueprint-spec). Trường cấu hình/plan đã được đối chiếu tài liệu, chưa được validate bằng API của tài khoản Render và chưa áp dụng thực tế.

### 4. Kiểm tra sau khi có URL thật

Thay `YOUR-SITE` bằng hostname mà nhà cung cấp cấp cho bạn, không dùng ví dụ như URL đã tồn tại:

```bash
curl -f https://YOUR-SITE/healthz
curl -f https://YOUR-SITE/readyz
curl -f https://YOUR-SITE/api/projects
```

Sau đó mở trình duyệt và kiểm tra:

- Trang chính hiện ảnh/CV/dự án đúng.
- `/readyz` trả `storage: postgresql`, không phải `memory`.
- Nút kiểm tra backend trả phản hồi thật.
- Lọc dự án, dark mode và màn hình điện thoại hoạt động.
- Gửi form hợp lệ rồi xác nhận có bản ghi trong database. Form **không gửi email**; không kiểm tra bằng việc chờ email đến hộp thư.
- Redeploy/restart web service, kiểm tra bản ghi vẫn còn; không dùng dữ liệu nhạy cảm cho bài test này.
- Không có endpoint public để đọc contact submissions; health/JSON không chứa secrets.

Với kết nối quản trị database được cấp quyền, truy vấn đọc tin nhắn:

```sql
SELECT id, name, email, message, created_at
FROM portfolio_contacts
ORDER BY created_at DESC
LIMIT 50;
```

Không đưa kết quả chứa email khách vào ảnh public. Nếu cần kết nối database ngoài Render, cấu hình IP allowlist tạm cho đúng IP của bạn rồi đóng lại; không mở toàn Internet chỉ cho tiện.

### 5. Domain riêng

- Không bắt buộc mua domain để nộp CV; URL HTTPS mặc định có thể dùng trước.
- Nếu có domain của bạn, cấu hình theo hướng dẫn nhà cung cấp rồi đặt `PUBLIC_URL=https://TEN-MIEN-THAT` (origin chính xác, không có path).
- Chọn một hostname chính và redirect hostname khác về đó. Nếu Origin không khớp `PUBLIC_URL`, form sẽ bị từ chối để bảo vệ CSRF.
- Không tự đổi link CV sang `nghi.dev` nếu bạn chưa sở hữu và deploy domain đó.

## Nếu dùng VPS

Cần VPS do bạn có quyền quản trị, DNS/domain hoặc phương án HTTPS, Docker và PostgreSQL. Compose đi kèm chỉ bind localhost, dành cho phát triển. Khi public cần reverse proxy HTTPS, `APP_ENV=production`, database thực, quản lý secrets, backup, firewall và kiểm tra rollback.

Không dùng IP/cổng của công ty cũ làm server portfolio nếu chưa được phép. Không gửi SSH private key/mật khẩu vào chat. Có thể tiếp tục hướng dẫn bằng cấu hình đã ẩn secrets và log đã làm sạch.

## Trước khi gắn link vào CV

Chỉ thêm mục Portfolio sau khi đã có URL thật và kiểm tra truy cập khi chưa đăng nhập. Không ghi "production Go experience" chỉ vì portfolio vừa được deploy; mô tả đúng đây là dự án cá nhân đã triển khai. Giữ source public có README/test để người phỏng vấn có thể kiểm tra và bạn giải thích được các quyết định kỹ thuật.
