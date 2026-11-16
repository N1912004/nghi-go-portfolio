(() => {
  'use strict';
  const root = document.documentElement;

  // Theme Management
  const themeToggle = document.querySelector('#theme-toggle');
  let savedTheme = null;
  try { savedTheme = localStorage.getItem('nghi-theme'); } catch (_) { /* Private browsing */ }
  const currentTheme = savedTheme === 'light' || savedTheme === 'dark' ? savedTheme : (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
  function setTheme(value) {
    root.dataset.theme = value;
    if (themeToggle) {
      themeToggle.setAttribute('aria-label', `Switch to ${value === 'dark' ? 'light' : 'dark'} theme`);
    }
    const meta = document.querySelector('meta[name="theme-color"]');
    if (meta) {
      meta.content = value === 'dark' ? '#101b21' : '#f7f8fa';
    }
  }
  setTheme(currentTheme);
  if (themeToggle) {
    themeToggle.hidden = false;
    themeToggle.addEventListener('click', () => {
      const value = root.dataset.theme === 'dark' ? 'light' : 'dark';
      setTheme(value);
      try { localStorage.setItem('nghi-theme', value); } catch (_) {}
    });
  }

  // Internationalization (i18n)
  const i18n = {
    "nav.about": { en: "About", vi: "Về tôi" },
    "nav.work": { en: "Work", vi: "Dự án" },
    "nav.engineering": { en: "Engineering", vi: "Kỹ thuật" },
    "nav.contact": { en: "Contact", vi: "Liên hệ" },
    "hero.eyebrow": { en: "BACKEND DEVELOPER · HO CHI MINH CITY", vi: "LẬP TRÌNH VIÊN BACKEND · TP. HỒ CHÍ MINH" },
    "hero.title": { en: "Good systems.<br>Built from <span class=\"accent\">the backend.</span>", vi: "Hệ thống chuẩn chỉnh.<br>Xây dựng từ <span class=\"accent\">backend.</span>" },
    "hero.intro": { en: "Hi, I'm <strong>Nguyen Quoc Nghi</strong>.<br>I build backend systems where the details matter — payments, data integrity, and reliable APIs.", vi: "Xin chào, tôi là <strong>Nguyễn Quốc Nghi</strong>.<br>Tôi phát triển các hệ thống backend chú trọng từng chi tiết — thanh toán, tính toàn vẹn dữ liệu và API tin cậy." },
    "hero.actions_work": { en: "Explore my work <span aria-hidden=\"true\">↗</span>", vi: "Xem dự án <span aria-hidden=\"true\">↗</span>" },
    "hero.actions_cv": { en: "Download CV <span aria-hidden=\"true\">↓</span>", vi: "Tải CV <span aria-hidden=\"true\">↓</span>" },
    "hero.foot": { en: "Production APIs in NestJS / Laravel.<br>Personal projects in Go.", vi: "API production với NestJS / Laravel.<br>Dự án cá nhân với Go." },
    "hero.role": { en: "Backend Developer", vi: "Lập trình viên Backend" },
    "hero.code_comment": { en: "// Build. Test. Understand.", vi: "// Xây dựng. Kiểm thử. Thấu hiểu." },
    "strip.tools": { en: "TOOLS I WORK WITH", vi: "CÔNG NGHỆ SỬ DỤNG" },
    "about.eyebrow": { en: "01 / THE PERSON BEHIND THE CODE", vi: "01 / CON NGƯỜI ĐẰNG SAU DÒNG CODE" },
    "about.title": { en: "Practical work.<br>Curious mindset.", vi: "Làm việc thực tế.<br>Tư duy tò mò." },
    "about.lead": { en: "At CIT Technology Solutions, I work on production APIs with NestJS and Laravel: payment reconciliation, concurrent wallet operations, database schemas, and query optimization. I also build personal projects in Go, exploring asynchronous jobs, worker pools, testing, and container delivery.", vi: "Tại CIT Technology Solutions, tôi phát triển các API production với NestJS và Laravel: đối soát thanh toán, xử lý đồng thời cho ví điện tử, thiết kế database schema và tối ưu truy vấn SQL. Tôi cũng tự xây dựng các dự án với Go, tìm hiểu về xử lý bất đồng bộ, worker pool, kiểm thử và đóng gói container." },
    "about.sub": { en: "I care about understanding failure cases, not just the happy path: duplicate payment callbacks, competing wallet updates, slow reports, and jobs that take longer than an HTTP request.", vi: "Tôi chú trọng vào việc tìm hiểu các tình huống lỗi thay vì chỉ luồng thuận lợi: trùng lặp callback thanh toán, cạnh tranh cập nhật số dư ví, báo cáo chậm và các tác vụ mất nhiều thời gian hơn một HTTP request thông thường." },
    "about.github": { en: "Explore my GitHub <span aria-hidden=\"true\">↗</span>", vi: "Khám phá GitHub của tôi <span aria-hidden=\"true\">↗</span>" },
    "about.prod_title": { en: "Production stack", vi: "Công nghệ production" },
    "about.go_title": { en: "Go & delivery practice", vi: "Thực hành Go & triển khai" },
    "exp.eyebrow": { en: "02 / EXPERIENCE", vi: "02 / KINH NGHIỆM LÀM VIỆC" },
    "exp.title": { en: "Shipping in<br>real environments.", vi: "Triển khai trong<br>môi trường thực tế." },
    "exp.period": { en: "Dec 2025 — Present", vi: "Tháng 12/2025 — Hiện tại" },
    "exp.role": { en: "Backend Developer", vi: "Lập trình viên Backend" },
    "exp.summary": { en: "Production backend development across payment, ride-hailing, ERP, and POS applications. Selected contributions are detailed below; Go work is listed separately as personal projects.", vi: "Phát triển backend cho các ứng dụng thực tế về thanh toán, gọi xe, ERP và hệ thống POS. Chi tiết các đóng góp tiêu biểu bên dưới; các dự án Go được liệt kê riêng." },
    "work.eyebrow": { en: "03 / SELECTED WORK", vi: "03 / DỰ ÁN TIÊU BIỂU" },
    "work.title": { en: "Problems I've worked on.", vi: "Những bài toán tôi đã giải quyết." },
    "work.note": { en: "My contributions, not just a list of technologies.<br>Client projects are described without private source code.", vi: "Đóng góp thực tế của tôi, không chỉ là danh sách công nghệ.<br>Các dự án khách hàng được mô tả mà không tiết lộ mã nguồn bảo mật." },
    "filter.all": { en: "All work", vi: "Tất cả" },
    "filter.go": { en: "Personal / Go", vi: "Cá nhân / Go" },
    "filter.prod": { en: "Production", vi: "Production" },
    "proj.contribution": { en: "My contribution", vi: "Đóng góp của tôi" },
    "proj.view_repo": { en: "Live website <span aria-hidden=\"true\">↗</span>", vi: "Trải nghiệm website <span aria-hidden=\"true\">↗</span>" },
    "eng.eyebrow": { en: "04 / UNDER THE HOOD", vi: "04 / BÊN TRONG HỆ THỐNG" },
    "eng.title": { en: "A portfolio with<br>a real Go backend.", vi: "Một portfolio với<br>backend Go thực thụ." },
    "eng.note": { en: "Server-rendered HTML. Embedded assets. PostgreSQL.<br>No framework needed to serve the first page.", vi: "HTML render từ server. Asset đóng gói kèm nhị phân. PostgreSQL.<br>Không cần framework cồng kềnh để phục vụ trang đầu tiên." },
    "eng.small_title": { en: "Small by design.", vi: "Tinh gọn từ thiết kế." },
    "eng.small_desc": { en: "This site uses Go's standard HTTP server and templates. Projects come from a store interface; contact messages are validated server-side. PostgreSQL is required in production.", vi: "Trang này sử dụng HTTP server và template chuẩn của Go. Dữ liệu dự án lấy từ store interface; tin nhắn liên hệ được xác thực phía server. Cần PostgreSQL khi chạy production." },
    "eng.li1": { en: "Parameterized SQL and context timeouts", vi: "Truy vấn SQL tham số hóa và context timeout" },
    "eng.li2": { en: "CSRF tokens, request limits, and security headers", vi: "CSRF token, giới hạn kích thước request và security header" },
    "eng.li3": { en: "Health/readiness checks and graceful shutdown", vi: "Health/readiness check và tắt ứng dụng an toàn (graceful shutdown)" },
    "eng.li4": { en: "Docker packaging and automated Go checks", vi: "Đóng gói Docker và kiểm thử Go tự động" },
    "eng.check_btn": { en: "Check backend <span aria-hidden=\"true\">↗</span>", vi: "Kiểm tra backend <span aria-hidden=\"true\">↗</span>" },
    "eng.view_json": { en: "View projects JSON ↗", vi: "Xem JSON dự án ↗" },
    "learn.eyebrow": { en: "05 / CONTINUOUS LEARNING", vi: "05 / HỌC HỎI LIÊN TỤC" },
    "learn.title": { en: "Keep learning.<br>Keep building.", vi: "Không ngừng học hỏi.<br>Không ngừng xây dựng." },
    "learn.edu": { en: "B.Sc. Information Technology — Ho Chi Minh City University of Industry and Trade (HUIT), June 2026", vi: "Cử nhân Công nghệ Thông tin — Đại học Công Thương TP.HCM (HUIT), Tháng 6/2026" },
    "contact.eyebrow": { en: "06 / GET IN TOUCH", vi: "06 / LIÊN HỆ" },
    "contact.title": { en: "Let's talk<br><span class=\"accent\">backend.</span>", vi: "Cùng trao đổi về<br><span class=\"accent\">backend.</span>" },
    "contact.desc": { en: "Have a backend opportunity or a technical question?<br>I'd be happy to hear from you.", vi: "Bạn có cơ hội nghề nghiệp backend hoặc câu hỏi kỹ thuật?<br>Rất vui được kết nối cùng bạn." },
    "contact.privacy": { en: "The form stores your name, email, and message. Messages are scheduled for deletion after 30 days; cleanup runs daily while the server is active. No advertising trackers.", vi: "Biểu mẫu lưu tên, email và tin nhắn của bạn. Tin nhắn được lên lịch tự động xóa sau 30 ngày; tiến trình dọn dẹp chạy hàng ngày khi server hoạt động. Không dùng mã theo dõi quảng cáo." },
    "contact.label_name": { en: "Your name", vi: "Họ và tên" },
    "contact.label_email": { en: "Email address", vi: "Địa chỉ email" },
    "contact.label_message": { en: "Message", vi: "Lời nhắn" },
    "contact.btn_send": { en: "Send message <span aria-hidden=\"true\">↗</span>", vi: "Gửi tin nhắn <span aria-hidden=\"true\">↗</span>" },
    "contact.note": { en: "20–2,000 characters. Please don't include passwords or confidential data.", vi: "Từ 20–2.000 ký tự. Vui lòng không đính kèm mật khẩu hoặc thông tin bí mật." },
    "footer.top": { en: "Back to top ↑", vi: "Lên đầu trang ↑" }
  };

  const placeholders = {
    "contact.ph_name": { en: "Your name", vi: "Họ và tên của bạn" },
    "contact.ph_email": { en: "you@company.com", vi: "email@example.com" },
    "contact.ph_message": { en: "Tell me about the opportunity or what you'd like to discuss…", vi: "Chia sẻ về cơ hội hợp tác hoặc nội dung bạn muốn trao đổi…" }
  };

  const projectsI18n = {
    "ai-gateway": {
      summary: {
        en: "An asynchronous gateway for long-running AI jobs, with bounded concurrency and provider fallback.",
        vi: "Cổng xử lý tác vụ AI bất đồng bộ cho các job chạy lâu, kiểm soát số lượng luồng đồng thời và tự động chuyển đổi dự phòng."
      },
      period: { en: "Oct 2026", vi: "Tháng 10/2026" },
      category: { en: "Personal / Go", vi: "Cá nhân / Go" },
      highlights: {
        en: [
          "HTTP 202 submission, polling/SSE results, bounded workers, per-job timeouts, and graceful shutdown.",
          "Atomic Redis Lua rate limiting, device-scoped job access, and interchangeable stores.",
          "15 tests run with the race detector; CI checks and Kubernetes probes/HPA configuration."
        ],
        vi: [
          "Xử lý tiếp nhận job HTTP 202, cơ chế polling/SSE nhận kết quả, worker pool giới hạn, timeout từng job và graceful shutdown.",
          "Giới hạn tần suất nguyên tử bằng Redis Lua script, phân quyền job theo thiết bị và cấu trúc store linh hoạt.",
          "15 bộ kiểm thử chạy với cờ race detector; CI checks và cấu hình Kubernetes probe/HPA."
        ]
      }
    },
    "unicorn": {
      summary: {
        en: "Backend for a cyber-cafe chain: PC-session billing, wallet top-ups, F&B orders, and revenue reports.",
        vi: "Hệ thống backend cho chuỗi cyber cafe: tính cước phiên máy PC, nạp ví, gọi món F&B và báo cáo doanh thu."
      },
      period: { en: "Mar 2026 — Present", vi: "Tháng 03/2026 — Hiện tại" },
      category: { en: "Production", vi: "Production" },
      highlights: {
        en: [
          "Unified payment IPN/callback handling with idempotency and per-minute reconciliation.",
          "Implemented multi-device admin sessions with Jest e2e tests.",
          "Tuned slow reporting queries with EXPLAIN ANALYZE."
        ],
        vi: [
          "Xử lý thống nhất callback/IPN thanh toán với cơ chế idempotency và đối soát định kỳ mỗi phút.",
          "Hiện thực phiên đăng nhập admin đa thiết bị kèm kiểm thử Jest e2e.",
          "Tối ưu hóa các truy vấn báo cáo chậm bằng EXPLAIN ANALYZE."
        ]
      }
    },
    "happygo": {
      summary: {
        en: "CMS and mobile APIs for a ride-hailing and delivery platform.",
        vi: "CMS và hệ thống REST API mobile cho nền tảng gọi xe và giao hàng."
      },
      period: { en: "Aug 2026 — Present", vi: "Tháng 08/2026 — Hiện tại" },
      category: { en: "Production", vi: "Production" },
      highlights: {
        en: [
          "Resolved approximately 60 tickets across trips, coupons, wallets, and stores.",
          "Added concurrency protection against duplicate driver-withdrawal approvals.",
          "Implemented Vietnamese search without diacritics and province-scoped access."
        ],
        vi: [
          "Giải quyết khoảng 60 ticket về chuyến đi, mã giảm giá, ví tài xế và cửa hàng.",
          "Thêm cơ chế chống duplicate khi duyệt lệnh rút tiền của tài xế.",
          "Hiện thực tìm kiếm tiếng Việt không dấu và phân quyền dữ liệu theo tỉnh thành."
        ]
      }
    },
    "catering-erp": {
      summary: {
        en: "Data models and ERP workflows for an industrial-catering business.",
        vi: "Mô hình dữ liệu và quy trình ERP cho doanh nghiệp cung cấp suất ăn công nghiệp."
      },
      period: { en: "Jul — Aug 2026", vi: "Tháng 07 — 08/2026" },
      category: { en: "Production", vi: "Production" },
      highlights: {
        en: [
          "Designed schemas and modules for HR/timekeeping, menus, purchasing, stock, supplier pricing, and food-safety checks.",
          "Added RBAC, Vietnamese/English localization, and Excel import/export.",
          "Improved query structure through N+1 removal and indexing."
        ],
        vi: [
          "Thiết kế schema và các phân hệ chấm công/nhân sự, thực đơn, thu mua, kho, định giá nhà cung cấp và kiểm soát an toàn thực phẩm.",
          "Thêm phân quyền RBAC, đa ngôn ngữ Việt/Anh và import/export Excel.",
          "Cải thiện hiệu năng truy vấn thông qua việc loại bỏ N+1 query và đánh chỉ mục."
        ]
      }
    },
    "posnail": {
      summary: {
        en: "Address-data migration for a US nail-salon POS management system.",
        vi: "Chuyển đổi và chuẩn hóa dữ liệu địa chỉ cho phần mềm quản lý salon nail tại Mỹ."
      },
      period: { en: "Aug — Sep 2026", vi: "Tháng 08 — 09/2026" },
      category: { en: "Production", vi: "Production" },
      highlights: {
        en: [
          "Replaced free-text addresses with indexed state/city IDs using a 29,880-city dataset.",
          "Preserved data across order, export, and audit-log flows."
        ],
        vi: [
          "Thay thế địa chỉ dạng văn bản tự do bằng ID bang/thành phố chuẩn với tập dữ liệu 29.880 thành phố.",
          "Đảm bảo tính toàn vẹn dữ liệu qua các luồng đặt hàng, xuất dữ liệu và nhật ký kiểm toán."
        ]
      }
    }
  };

  const trainingI18n = [
    {
      title: { en: "Learning Docker", vi: "Khóa học Docker thực tế" },
      sub: { en: "LinkedIn Learning · Mar 2026 · Completed training", vi: "LinkedIn Learning · Tháng 03/2026 · Đã hoàn thành" }
    },
    {
      title: { en: "Practical GitHub Actions", vi: "Khóa học GitHub Actions thực hành" },
      sub: { en: "LinkedIn Learning · Mar 2026 · Completed training", vi: "LinkedIn Learning · Tháng 03/2026 · Đã hoàn thành" }
    },
    {
      title: { en: "Career Essentials in GitHub Copilot Professional Certificate", vi: "Chứng chỉ Chuyên nghiệp GitHub Copilot" },
      sub: { en: "LinkedIn Learning · Mar 2026 · Completed training", vi: "LinkedIn Learning · Tháng 03/2026 · Đã hoàn thành" }
    }
  ];

  const langToggle = document.querySelector('#lang-toggle');
  let currentLang = 'en';
  try {
    const saved = localStorage.getItem('nghi-lang');
    if (saved === 'en' || saved === 'vi') currentLang = saved;
  } catch (_) {}

  function setLanguage(lang) {
    currentLang = lang;
    root.lang = lang;
    if (langToggle) {
      langToggle.textContent = lang === 'vi' ? 'EN' : 'VI';
      langToggle.setAttribute('aria-label', lang === 'vi' ? 'Switch to English' : 'Chuyển sang tiếng Việt');
    }

    // Translate general elements
    document.querySelectorAll('[data-i18n]').forEach(el => {
      const key = el.dataset.i18n;
      if (i18n[key] && i18n[key][lang]) {
        el.innerHTML = i18n[key][lang];
      }
    });

    // Translate input placeholders
    document.querySelectorAll('[data-i18n-ph]').forEach(el => {
      const key = el.dataset.i18nPh;
      if (placeholders[key] && placeholders[key][lang]) {
        el.placeholder = placeholders[key][lang];
      }
    });

    // Translate project cards
    Object.keys(projectsI18n).forEach(id => {
      const p = projectsI18n[id];
      const summaryEl = document.querySelector(`[data-summary-id="${id}"]`);
      if (summaryEl && p.summary && p.summary[lang]) {
        summaryEl.textContent = p.summary[lang];
      }
      const periodEl = document.querySelector(`[data-period-id="${id}"]`);
      if (periodEl && p.period && p.period[lang]) {
        periodEl.textContent = p.period[lang];
      }
      const catEl = document.querySelector(`[data-category-badge="${id}"]`);
      if (catEl && p.category && p.category[lang]) {
        catEl.textContent = p.category[lang];
      }
      const highlightsEl = document.querySelector(`[data-highlights-id="${id}"]`);
      if (highlightsEl && p.highlights && p.highlights[lang]) {
        highlightsEl.innerHTML = p.highlights[lang].map(h => `<li>${h}</li>`).join('');
      }
    });

    // Translate training cards
    trainingI18n.forEach((t, i) => {
      const titleEl = document.querySelector(`[data-training-title="${i}"]`);
      if (titleEl && t.title && t.title[lang]) {
        titleEl.textContent = t.title[lang];
      }
      const subEl = document.querySelector(`[data-training-sub="${i}"]`);
      if (subEl && t.sub && t.sub[lang]) {
        subEl.textContent = t.sub[lang];
      }
    });
  }

  // Initial language setup
  setLanguage(currentLang);

  if (langToggle) {
    langToggle.addEventListener('click', () => {
      const next = currentLang === 'vi' ? 'en' : 'vi';
      setLanguage(next);
      try { localStorage.setItem('nghi-lang', next); } catch (_) {}
    });
  }

  // Project Filtering
  const filters = document.querySelector('#project-filters');
  const cards = [...document.querySelectorAll('.project-card')];
  if (filters) {
    filters.hidden = false;
    filters.addEventListener('click', event => {
      const button = event.target.closest('button[data-filter]');
      if (!button) return;
      filters.querySelectorAll('button').forEach(item => {
        const active = item === button;
        item.classList.toggle('active', active);
        item.setAttribute('aria-pressed', String(active));
      });
      let count = 0;
      cards.forEach(card => {
        card.hidden = button.dataset.filter !== 'all' && card.dataset.category !== button.dataset.filter;
        if (!card.hidden) count++;
      });
      const filterStatus = document.querySelector('#filter-status');
      if (filterStatus) {
        filterStatus.textContent = currentLang === 'vi' ? `Hiển thị ${count} dự án` : `${count} projects shown`;
      }
    });
  }

  // Backend Health/Readiness Check
  const apiActions = document.querySelector('.api-actions');
  const check = document.querySelector('#check-backend');
  if (apiActions && check) {
    apiActions.hidden = false;
    check.addEventListener('click', async () => {
      const status = document.querySelector('#backend-status');
      const output = document.querySelector('#api-output');
      check.disabled = true;
      if (status) {
        status.textContent = currentLang === 'vi' ? 'Đang kiểm tra backend…' : 'Checking the running backend…';
      }
      const started = performance.now();
      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), 6000);
      try {
        const response = await fetch('/readyz', { signal: controller.signal, cache: 'no-store' });
        const body = await response.json();
        const elapsed = Math.round(performance.now() - started);
        if (output) {
          output.textContent = `GET /readyz → HTTP ${response.status}\n\n${JSON.stringify(body, null, 2)}`;
        }
        if (status) {
          status.textContent = response.ok
            ? (currentLang === 'vi' ? `Backend đã sẵn sàng · phản hồi trong ${elapsed} ms.` : `Backend ready · ${elapsed} ms for this request (not a benchmark).`)
            : (currentLang === 'vi' ? 'Kiểm tra thất bại. Vui lòng thử lại sau.' : 'Readiness check failed. Please try again later.');
        }
      } catch (_) {
        if (output) {
          output.textContent = currentLang === 'vi' ? 'Không thể kết nối tới backend.' : 'The backend could not be reached. No response has been simulated.';
        }
        if (status) {
          status.textContent = currentLang === 'vi' ? 'Kết nối không khả dụng. Vui lòng thử lại.' : 'Connection unavailable. Please try again.';
        }
      } finally {
        clearTimeout(timeout);
        check.disabled = false;
      }
    });
  }
})();