// Initialize Lucide icons
lucide.createIcons();

// Initialize Lucide icons when DOM loads
document.addEventListener("DOMContentLoaded", function () {
  // Initialize tooltips if Bootstrap is available
  if (typeof bootstrap !== "undefined") {
    const tooltipTriggerList = [].slice.call(
      document.querySelectorAll("[title]")
    );
    tooltipTriggerList.map(function (tooltipTriggerEl) {
      return new bootstrap.Tooltip(tooltipTriggerEl);
    });
  }

  // Auto-hide alerts after 5 seconds
  setTimeout(function () {
    const alerts = document.querySelectorAll(".alert.show");
    alerts.forEach(function (alert) {
      const bsAlert = new bootstrap.Alert(alert);
      bsAlert.close();
    });
  }, 5000);

  // Add loading states to buttons
  const buttons = document.querySelectorAll(".btn");
  buttons.forEach(function (button) {
    button.addEventListener("click", function () {
      if (
        this.type === "submit" ||
        this.classList.contains("btn-primary-custom")
      ) {
        this.disabled = true;
        const originalText = this.innerHTML;
        this.innerHTML =
          '<i data-lucide="loader-2" class="me-2 animate-spin"></i>Processando...';

        // Re-enable after 3 seconds if form is not submitted
        setTimeout(() => {
          if (this.disabled) {
            this.disabled = false;
            this.innerHTML = originalText;
            lucide.createIcons();
          }
        }, 3000);
      }
    });
  });

  // Validação customizada para o campo service_type
  const form = document.querySelector("form");
  if (form) {
    form.addEventListener("submit", function (e) {
      const radios = document.querySelectorAll('input[name="service_type"]');
      const checked = Array.from(radios).some((r) => r.checked);

      if (!checked) {
        e.preventDefault(); // bloqueia envio
        showNotification("Por favor, selecione o tipo de serviço.", "warning");
      }
    });
  }
});

// Função para cancelar solicitação
function cancelarSolicitacao(id) {
  const form = document.getElementById("cancelForm");
  if (form) {
    form.action = "/solicitacao/" + id + "/cancelar";

    if (typeof bootstrap !== "undefined") {
      const modal = new bootstrap.Modal(document.getElementById("cancelModal"));
      modal.show();
    }
  }
}

// Function to refresh page data (can be called by admin updates)
function refreshData() {
  window.location.reload();
}

// Function to format dates nicely
function formatDate(dateString) {
  const date = new Date(dateString);
  const options = {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  };
  return date.toLocaleDateString("pt-BR", options);
}

// Function to get status color class
function getStatusColorClass(status) {
  const statusMap = {
    SOLICITADA: "status-solicitada",
    CONFIRMADA: "status-confirmada",
    REALIZADA: "status-realizada",
    CANCELADA: "status-cancelada",
    // Legacy support
    pendente: "status-pendente",
    em_andamento: "status-em_andamento",
    concluido: "status-concluido",
  };

  return statusMap[status] || "status-solicitada";
}

// Function to get status display text
function getStatusDisplayText(status) {
  const statusMap = {
    SOLICITADA: "Solicitada",
    CONFIRMADA: "Confirmada",
    REALIZADA: "Realizada",
    CANCELADA: "Cancelada",
    // Legacy support
    pendente: "Pendente",
    em_andamento: "Em Andamento",
    concluido: "Concluído",
  };

  return statusMap[status] || "Desconhecido";
}

// Function to show notification
function showNotification(message, type = "info") {
  const alertClass = `alert-${type}`;
  const iconMap = {
    success: "check-circle",
    error: "x-circle",
    warning: "alert-triangle",
    info: "info",
  };

  const alert = document.createElement("div");
  alert.className = `alert ${alertClass} alert-dismissible fade show position-fixed`;
  alert.style.cssText =
    "top: 20px; right: 20px; z-index: 9999; min-width: 300px;";
  alert.innerHTML = `
        <i data-lucide="${iconMap[type]}" class="me-2"></i>
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;

  document.body.appendChild(alert);

  // Initialize Lucide icons for the new alert
  lucide.createIcons();

  // Auto remove after 5 seconds
  setTimeout(() => {
    if (alert.parentNode) {
      const bsAlert = new bootstrap.Alert(alert);
      bsAlert.close();
    }
  }, 5000);
}

// Function to confirm action with custom message
function confirmAction(message, callback) {
  if (confirm(message)) {
    callback();
  }
}

// Function to handle table responsive behavior
function handleTableResponsive() {
  const tables = document.querySelectorAll(".table-responsive");
  tables.forEach(function (table) {
    // Add swipe hint for mobile
    if (window.innerWidth <= 768 && table.scrollWidth > table.clientWidth) {
      if (!table.querySelector(".swipe-hint")) {
        const hint = document.createElement("div");
        hint.className = "swipe-hint text-muted small text-center mt-2";
        hint.innerHTML =
          '<i data-lucide="chevrons-right" class="me-1"></i>Deslize para ver mais';
        table.parentNode.insertBefore(hint, table.nextSibling);
        lucide.createIcons();
      }
    }
  });
}

// Handle window resize
window.addEventListener("resize", handleTableResponsive);

// Function to copy text to clipboard
function copyToClipboard(text) {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(text).then(function () {
      showNotification("Texto copiado para a área de transferência!", "success");
    });
  } else {
    // Fallback for older browsers
    const textArea = document.createElement("textarea");
    textArea.value = text;
    document.body.appendChild(textArea);
    textArea.select();
    document.execCommand("copy");
    document.body.removeChild(textArea);
    showNotification("Texto copiado para a área de transferência!", "success");
  }
}
