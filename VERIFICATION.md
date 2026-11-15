# Kết quả kiểm tra portfolio Go

## Đã xác minh trong môi trường hiện tại

- **Build:** biên dịch được binary Go với `go build -mod=readonly -trimpath`; toolchain Go 1.27.4.
- **Static checks:** `go vet ./...` hoàn thành không báo lỗi.
- **Tests:** 26 test cấp cao pass, 0 fail; chạy với `go test -race -count=1`. Một test PostgreSQL integration skip do chưa có `TEST_DATABASE_URL`.
- **Coverage:** 53,9% statements cho lượt test local. Chỉ phản ánh nhánh đã chạy, không phải điểm chất lượng/bảo mật và không phải coverage PostgreSQL production.
- **Rendering:** server Go thật phục vụ trang, ảnh và CV PDF; nội dung profile/dự án có sẵn khi tắt JavaScript.
- **Browser:** Chromium headless kiểm tra giao diện 1.440×1.000, điện thoại 390×844 và màn hình hẹp 320×740. Không thấy tràn ngang ở hai độ rộng mobile đã kiểm tra.
- **Tương tác:** bộ lọc hiển thị lần lượt 5 dự án / 1 dự án cá nhân Go / 4 production; phần đóng góp mở được; đổi theme và lưu lựa chọn hoạt động.
- **Backend:** nút Check backend gọi `/readyz` thật và hiển thị JSON, không mô phỏng trạng thái. Local trả rõ `storage: memory`.
- **Contact:** trình duyệt gửi form có CSRF cookie/token, lưu vào bộ nhớ development và nhận phản hồi đúng. Không gửi email ra ngoài.
- **Trình duyệt:** không ghi nhận lỗi JavaScript/console trong các thao tác đã kiểm tra.
- **Biện pháp HTTP:** tests kiểm tra security headers, CSRF thiếu/sai, Origin khác, honeypot, validation, body quá lớn, quota, template escaping, failure response và startup production thiếu database.

Các ảnh preview được chụp từ trang do binary Go local phục vụ, không phải mockup hoặc ảnh template trên Internet.

## Chưa được xác minh

- **PostgreSQL thật:** chưa có database được cấp để chạy end-to-end. Source có pgx store, tạo/seed schema trong transaction, parameterized queries, readiness ping, contact retention và integration test schema tạm; phải chạy integration trước khi vận hành production.
- **Docker:** chưa chạy build/run container do môi trường không có Docker daemon. Workflow GitHub có bước Docker build, nhưng workflow chưa được chạy trên repo người dùng.
- **GitHub CI:** mới có cấu hình, chưa có một run GitHub Actions thực tế để trỏ badge hoặc tuyên bố CI passing.
- **Cloud:** chưa deploy Render/Cloud Run/VPS, chưa áp dụng Blueprint, chưa tạo domain, chưa tạo tài nguyên trả phí. Không có link live để đưa vào CV.
- **Load/security audit:** chưa benchmark, pentest, test đa replica hoặc kiểm thử hết trình duyệt/thiết bị. Tests cơ bản không phải chứng nhận bảo mật.
- **AI gateway:** website chỉ liên kết repo đã nêu trong CV; chưa kết nối gateway, chưa gọi nhà cung cấp AI, không có demo tác vụ AI giả làm thật.

## Giới hạn cần biết trước khi publish

- Form lưu tin, không gửi email thông báo; chủ website cần đọc DB hoặc bổ sung kênh thông báo được cấu hình và kiểm thử riêng.
- Memory mode chỉ dùng phát triển. Production từ chối khởi động nếu thiếu database hoặc origin HTTPS.
- Quota liên hệ lưu RAM theo IP kết nối, không distributed; khi có proxy, nhiều khách có thể dùng chung quota. Cần rate limit ở edge hoặc cấu hình shared limiter cho nhu cầu lớn hơn.
- Nội dung dự án/ảnh/CV được lấy từ hồ sơ đã cung cấp; chưa xác minh độc lập thành tích tại công ty hoặc quyền công khai tên dự án. Không đính kèm source công ty.
- Logo `nghi.dev` là nhãn thiết kế, không khẳng định sở hữu domain. Ảnh và CV chứa thông tin cá nhân thật: kiểm tra lại trước khi công khai.
