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

// Form validation
document.querySelector("form").addEventListener("submit", function (e) {
  const serviceType = document.querySelector(
    'input[name="service_type"]:checked'
  );
  const description = document.querySelector("#description").value.trim();

  if (!serviceType) {
    e.preventDefault();
    alert("Por favor, selecione um tipo de serviço.");
    return;
  }

  if (description.length < 20) {
    e.preventDefault();
    alert(
      "Por favor, forneça uma descrição mais detalhada (mínimo 20 caracteres)."
    );
    return;
  }
});
