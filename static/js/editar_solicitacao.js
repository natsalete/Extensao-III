// Initialize Select2
$(document).ready(function () {
  $(".select2").select2({
    placeholder: "Selecione um estado",
    allowClear: false,
    language: {
      noResults: function () {
        return "Nenhum resultado encontrado";
      },
      searching: function () {
        return "Buscando...";
      },
    },
  });
});

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

    if (value.replace("-", "").length === 8) {
      buscarCEP(value);
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
      mostrarNotificacao("CEP não encontrado", "warning");
      return;
    }

    if (data.logradouro)
      document.getElementById("logradouro").value = data.logradouro;
    if (data.bairro) document.getElementById("bairro").value = data.bairro;
    if (data.localidade)
      document.getElementById("cidade").value = data.localidade;
    if (data.uf) {
      $("#estado").val(data.uf).trigger("change");
    }

    document.getElementById("numero").focus();
    mostrarNotificacao("Endereço preenchido com sucesso!", "success");
  } catch (error) {
    console.error("Erro ao buscar CEP:", error);
    mostrarNotificacao("Erro ao buscar CEP. Tente novamente.", "error");
  } finally {
    if (cepLoading) cepLoading.classList.add("d-none");
    if (searchBtn) searchBtn.disabled = false;
    if (cepInputField) cepInputField.disabled = false;
  }
}

// Set minimum date to today
const dateInput = document.getElementById("preferred_date");
if (dateInput) {
  const today = new Date().toISOString().split("T")[0];
  dateInput.setAttribute("min", today);
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
      mostrarNotificacao("Selecione um tipo de serviço.", "warning");
      return false;
    }

    const fullName = document.getElementById("full_name").value.trim();
    if (fullName.length < 3) {
      e.preventDefault();
      mostrarNotificacao("Digite seu nome completo.", "warning");
      return false;
    }

    const cep = document.getElementById("cep").value.replace(/\D/g, "");
    if (cep.length !== 8) {
      e.preventDefault();
      mostrarNotificacao("Digite um CEP válido.", "warning");
      return false;
    }

    const submitBtn = this.querySelector('button[type="submit"]');
    if (submitBtn) {
      submitBtn.disabled = true;
      submitBtn.innerHTML =
        '<i class="bi bi-arrow-repeat animate-spin me-2"></i>Salvando...';
    }

    return true;
  });
}

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

  const alert = document.createElement("div");
  alert.className = `alert ${alertClass} alert-dismissible fade show position-fixed`;
  alert.style.cssText =
    "top: 20px; right: 20px; z-index: 9999; min-width: 300px;";
  alert.innerHTML = `
            <div class="d-flex align-items-center">
                <i class="bi bi-${iconMap[tipo]} me-2"></i>
                <div class="flex-grow-1">${mensagem}</div>
                <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
            </div>
        `;

  document.body.appendChild(alert);

  setTimeout(() => {
    if (alert.parentNode) {
      bootstrap.Alert.getOrCreateInstance(alert).close();
    }
  }, 5000);
}
