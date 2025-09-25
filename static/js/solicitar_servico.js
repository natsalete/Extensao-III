// Initialize Lucide icons
lucide.createIcons();

// Handle service option selection
document.querySelectorAll(".service-option").forEach((option) => {
    option.addEventListener("click", function () {
        // Remove selected class from all options
        document.querySelectorAll(".service-option").forEach((opt) => {
            opt.classList.remove("selected");
        });

        // Add selected class to clicked option
        this.classList.add("selected");

        // Check the radio button
        const radio = this.querySelector('input[type="radio"]');
        if (radio) {
            radio.checked = true;
        }
    });
});

// CEP mask and validation
document.getElementById('cep').addEventListener('input', function (e) {
    let value = e.target.value.replace(/\D/g, '');
    value = value.replace(/^(\d{5})(\d)/, '$1-$2');
    e.target.value = value;
    
    if (value.length === 9) {
        buscarCEP(value);
    }
});

// Buscar CEP
function buscarCEP(cep) {
    const cepLimpo = cep.replace('-', '');
    
    if (cepLimpo.length === 8) {
        fetch(`https://viacep.com.br/ws/${cepLimpo}/json/`)
            .then(response => response.json())
            .then(data => {
                if (!data.erro) {
                    document.getElementById('logradouro').value = data.logradouro || document.getElementById('logradouro').value;
                    document.getElementById('bairro').value = data.bairro || document.getElementById('bairro').value;
                    document.getElementById('cidade').value = data.localidade || document.getElementById('cidade').value;
                    document.getElementById('estado').value = data.uf || document.getElementById('estado').value;
                }
            })
            .catch(error => console.log('Erro ao buscar CEP:', error));
    }
}

// Set minimum date to today
const today = new Date().toISOString().split('T')[0];
document.getElementById('preferred_date').setAttribute('min', today);

// Form validation
document.getElementById("serviceForm").addEventListener("submit", function (e) {
    const serviceType = document.querySelector('input[name="service_type"]:checked');
    const fullName = document.getElementById("full_name").value.trim();
    const cep = document.getElementById("cep").value.trim();
    const preferredDate = document.getElementById("preferred_date").value;

    if (!serviceType) {
        e.preventDefault();
        alert("Por favor, selecione um tipo de serviço.");
        return;
    }

    if (fullName.length < 3) {
        e.preventDefault();
        alert("Por favor, digite seu nome completo.");
        return;
    }

    if (cep.length !== 9) {
        e.preventDefault();
        alert("Por favor, digite um CEP válido.");
        return;
    }

    if (!preferredDate) {
        e.preventDefault();
        alert("Por favor, selecione uma data preferencial.");
        return;
    }
});
