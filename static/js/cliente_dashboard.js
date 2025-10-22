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

function showNotification(message, type = "info") {
  const alertClass = {
    success: 'alert-success',
    error: 'alert-danger',
    warning: 'alert-warning',
    info: 'alert-info'
  }[type] || 'alert-info';
  
  const iconMap = {
    success: "bi-check-circle-fill",
    error: "bi-x-circle-fill",
    warning: "bi-exclamation-triangle-fill",
    info: "bi-info-circle-fill",
  };

  const alert = document.createElement("div");
  alert.className = `alert ${alertClass} alert-dismissible fade show position-fixed`;
  alert.style.cssText = "top: 20px; right: 20px; z-index: 9999; min-width: 300px;";
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