// Handle service option selection
document.querySelectorAll(".service-option").forEach((option) => {
  option.addEventListener("click", function () {
    document.querySelectorAll(".service-option").forEach((opt) => {
      opt.classList.remove("selected");
    });
    this.classList.add("selected");

    const radio = this.querySelector('input[type="radio"]');
    if (radio) {
      radio.checked = true;
    }
  });
});

// CEP Auto-format
const cepInput = document.getElementById("cep");
if (cepInput) {
  cepInput.addEventListener("input", function (e) {
    let value = e.target.value.replace(/\D/g, "");

    if (value.length > 5) {
      value = value.substring(0, 5) + "-" + value.substring(5, 8);
    }

    e.target.value = value;
  });

  // Auto-buscar quando completar 8 dígitos
  cepInput.addEventListener("blur", function (e) {
    const cep = e.target.value.replace(/\D/g, "");
    if (cep.length === 8) {
      buscarCEP(e.target.value);
    }
  });
}

// Botão de buscar CEP
const searchCepBtn = document.getElementById("searchCepBtn");
if (searchCepBtn) {
  searchCepBtn.addEventListener("click", function (e) {
    e.preventDefault();
    const cep = document.getElementById("cep").value;
    if (cep) {
      buscarCEP(cep);
    }
  });
}

// Função para buscar CEP
async function buscarCEP(cep) {
  const cepLimpo = cep.replace(/\D/g, "");

  if (cepLimpo.length !== 8) {
    mostrarNotificacao("CEP deve conter 8 dígitos", "warning");
    return;
  }

  const cepLoading = document.getElementById("cepLoading");
  const searchBtn = document.getElementById("searchCepBtn");
  const cepInputField = document.getElementById("cep");

  // Mostrar loading
  if (cepLoading) cepLoading.classList.remove("d-none");
  if (searchBtn) searchBtn.disabled = true;
  if (cepInputField) cepInputField.disabled = true;

  try {
    const response = await fetch(`https://viacep.com.br/ws/${cepLimpo}/json/`);

    if (!response.ok) {
      throw new Error("Erro ao buscar CEP");
    }

    const data = await response.json();

    if (data.erro) {
      mostrarNotificacao("CEP não encontrado. Verifique e tente novamente.", "warning");
      return;
    }

    // Preencher campos apenas se vieram dados
    let camposPreenchidos = 0;
    
    if (data.logradouro) {
      document.getElementById("logradouro").value = data.logradouro;
      camposPreenchidos++;
    }
    if (data.bairro) {
      document.getElementById("bairro").value = data.bairro;
      camposPreenchidos++;
    }
    if (data.localidade) {
      document.getElementById("cidade").value = data.localidade;
      camposPreenchidos++;
    }
    if (data.uf) {
      document.getElementById("estado").value = data.uf;
      camposPreenchidos++;
    }

    // Mostrar mensagem de sucesso apenas se preencheu algum campo
    if (camposPreenchidos > 0) {
      document.getElementById("numero").focus();
      mostrarNotificacao("Endereço preenchido com sucesso!", "success");
    } else {
      mostrarNotificacao("CEP válido, mas sem dados de endereço. Preencha manualmente.", "info");
    }

  } catch (error) {
    console.error("Erro ao buscar CEP:", error);
    mostrarNotificacao("Erro ao buscar CEP. Verifique sua conexão e tente novamente.", "error");
  } finally {
    // Remover loading
    if (cepLoading) cepLoading.classList.add("d-none");
    if (searchBtn) searchBtn.disabled = false;
    if (cepInputField) cepInputField.disabled = false;
  }
}

// Set minimum date
const dateInput = document.getElementById("preferred_date");
if (dateInput) {
  const isEditPage = window.location.pathname.includes('/editar');
  
  if (isEditPage) {
    // Na página de edição, permite data de hoje
    const today = new Date().toISOString().split("T")[0];
    dateInput.setAttribute("min", today);
  } else {
    // Na página de nova solicitação, mínimo é amanhã
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    const minDate = tomorrow.toISOString().split("T")[0];
    dateInput.setAttribute("min", minDate);
    
    // Define valor padrão apenas se não tiver valor
    if (!dateInput.value) {
      dateInput.value = minDate;
    }
  }
}

// Set default time
const timeInput = document.getElementById("preferred_time");
if (timeInput && !timeInput.value) {
  timeInput.value = "09:00";
}

// Form validation
const serviceForm = document.getElementById("serviceForm");
if (serviceForm) {
  serviceForm.addEventListener("submit", function (e) {
    const serviceType = document.querySelector(
      'input[name="service_type"]:checked'
    );
    if (!serviceType) {
      e.preventDefault();
      mostrarNotificacao("Por favor, selecione um tipo de serviço.", "warning");
      return false;
    }

    const fullName = document.getElementById("full_name").value.trim();
    if (fullName.length < 3) {
      e.preventDefault();
      mostrarNotificacao("Por favor, digite seu nome completo.", "warning");
      document.getElementById("full_name").focus();
      return false;
    }

    const cep = document.getElementById("cep").value.replace(/\D/g, "");
    if (cep.length !== 8) {
      e.preventDefault();
      mostrarNotificacao("Por favor, digite um CEP válido com 8 dígitos.", "warning");
      document.getElementById("cep").focus();
      return false;
    }

    // Validar horário
    const time = document.getElementById("preferred_time").value;
    if (time) {
      const [hours, minutes] = time.split(':').map(Number);
      if (hours < 8 || hours >= 17) {
        e.preventDefault();
        mostrarNotificacao("Por favor, selecione um horário entre 08:00 e 17:00.", "warning");
        document.getElementById("preferred_time").focus();
        return false;
      }
    }

    const submitBtn = this.querySelector('button[type="submit"]');
    if (submitBtn) {
      submitBtn.disabled = true;
      const isEditPage = window.location.pathname.includes('/editar');
      const btnText = isEditPage ? 'Salvando...' : 'Enviando...';
      submitBtn.innerHTML = `<i class="bi bi-arrow-repeat animate-spin me-2"></i>${btnText}`;
    }

    return true;
  });
}

// Função para mostrar notificações
function mostrarNotificacao(mensagem, tipo = "info") {
  const alertClass =
    {
      success: "alert-success",
      error: "alert-danger",
      warning: "alert-warning",
      info: "alert-info",
    }[tipo] || "alert-info";

  const iconMap = {
    success: "check-circle-fill",
    error: "x-circle-fill",
    warning: "exclamation-triangle-fill",
    info: "info-circle-fill",
  };

  // Remover notificações anteriores
  const existingAlerts = document.querySelectorAll('.notification-alert');
  existingAlerts.forEach(alert => alert.remove());

  const alert = document.createElement("div");
  alert.className = `alert ${alertClass} alert-dismissible fade show position-fixed notification-alert`;
  alert.style.cssText =
    "top: 20px; right: 20px; z-index: 9999; min-width: 300px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);";
  alert.innerHTML = `
    <div class="d-flex align-items-center">
      <i class="bi bi-${iconMap[tipo]} me-2"></i>
      <div class="flex-grow-1">${mensagem}</div>
      <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    </div>
  `;

  document.body.appendChild(alert);

  // Auto-remover após 5 segundos
  setTimeout(() => {
    if (alert.parentNode) {
      alert.classList.remove('show');
      setTimeout(() => alert.remove(), 150);
    }
  }, 5000);
}

// Inicializar Lucide icons se estiver na página de edição
if (typeof lucide !== 'undefined') {
  lucide.createIcons();
}