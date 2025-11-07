document.addEventListener("DOMContentLoaded", function () {
  // Auto-fechar alertas após 5 segundos
  setTimeout(function () {
    const alerts = document.querySelectorAll(".alert.show");
    alerts.forEach(function (alert) {
      if (typeof bootstrap !== "undefined") {
        bootstrap.Alert.getOrCreateInstance(alert).close();
      }
    });
  }, 5000);

  // Inicializar tooltips se houver
  initializeTooltips();

  // Adicionar animação aos cards de estatísticas
  animateStatCards();
});

function cancelarSolicitacao(id) {
  const form = document.getElementById("cancelForm");
  if (form) {
    form.action = "/solicitacao/" + id + "/cancelar";

    if (typeof bootstrap !== "undefined") {
      const modalElement = document.getElementById("cancelModal");
      if (modalElement) {
        const modal = new bootstrap.Modal(modalElement);
        modal.show();
      }
    }
  }
}

function clearFilters() {
  window.location.href = "/dashboard/cliente";
}

function showNotification(message, type = "info") {
  const alertClass = {
    success: "alert-success",
    error: "alert-danger",
    warning: "alert-warning",
    info: "alert-info",
  }[type] || "alert-info";

  const iconMap = {
    success: "bi-check-circle-fill",
    error: "bi-x-circle-fill",
    warning: "bi-exclamation-triangle-fill",
    info: "bi-info-circle-fill",
  };

  const alert = document.createElement("div");
  alert.className = `alert ${alertClass} alert-dismissible fade show position-fixed`;
  alert.style.cssText =
    "top: 20px; right: 20px; z-index: 9999; min-width: 300px;";
  alert.innerHTML = `
    <div class="d-flex align-items-center">
      <i class="bi ${iconMap[type]} me-2"></i>
      <div class="flex-grow-1">${message}</div>
      <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    </div>
  `;

  document.body.appendChild(alert);

  setTimeout(() => {
    if (alert.parentNode && typeof bootstrap !== "undefined") {
      bootstrap.Alert.getOrCreateInstance(alert).close();
    }
  }, 5000);
}

function initializeTooltips() {
  if (typeof bootstrap !== "undefined") {
    const tooltipTriggerList = [].slice.call(
      document.querySelectorAll('[data-bs-toggle="tooltip"]')
    );
    tooltipTriggerList.map(function (tooltipTriggerEl) {
      return new bootstrap.Tooltip(tooltipTriggerEl);
    });
  }
}

function animateStatCards() {
  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.style.opacity = "0";
          entry.target.style.transform = "translateY(20px)";

          setTimeout(() => {
            entry.target.style.transition = "all 0.5s ease";
            entry.target.style.opacity = "1";
            entry.target.style.transform = "translateY(0)";
          }, 100);

          observer.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.1 }
  );

  document.querySelectorAll(".service-card").forEach((card, index) => {
    card.style.transitionDelay = `${index * 0.1}s`;
    observer.observe(card);
  });
}

// Função para highlight dos filtros ativos
function highlightActiveFilters() {
  const urlParams = new URLSearchParams(window.location.search);
  const statusFilter = urlParams.get("status");
  const serviceTypeFilter = urlParams.get("service_type");

  if (statusFilter || serviceTypeFilter) {
    const filterSection = document.querySelector(".filter-section");
    if (filterSection) {
      filterSection.classList.add("active-filters");
    }
  }
}

// Executar ao carregar
highlightActiveFilters();

// Prevenir múltiplos submits do formulário
document.addEventListener("submit", function (e) {
  const form = e.target;
  const submitButton = form.querySelector('button[type="submit"]');

  if (submitButton && !submitButton.disabled) {
    submitButton.disabled = true;
    submitButton.innerHTML =
      '<span class="spinner-border spinner-border-sm me-2"></span>Processando...';

    // Reabilitar após 3 segundos caso haja erro
    setTimeout(() => {
      submitButton.disabled = false;
      submitButton.innerHTML = submitButton.dataset.originalText || "Enviar";
    }, 3000);
  }
});

// Salvar texto original dos botões
document.querySelectorAll('button[type="submit"]').forEach((btn) => {
  btn.dataset.originalText = btn.innerHTML;
});