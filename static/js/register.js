// Máscara para telefone brasileiro
function applyPhoneMask(value) {
  // Remove tudo que não é dígito
  value = value.replace(/\D/g, "");

  // Aplica a máscara baseada no tamanho
  if (value.length <= 10) {
    // Formato: (99) 9999-9999
    value = value.replace(/(\d{2})(\d{4})(\d{4})/, "($1) $2-$3");
  } else {
    // Formato: (99) 99999-9999
    value = value.replace(/(\d{2})(\d{5})(\d{4})/, "($1) $2-$3");
  }

  return value;
}

// Aplica a máscara ao campo de telefone
document.getElementById("phone").addEventListener("input", function (e) {
  let value = e.target.value;

  // Remove caracteres não numéricos para verificar o tamanho
  let numbersOnly = value.replace(/\D/g, "");

  // Limita a 11 dígitos (celular com 9 dígitos)
  if (numbersOnly.length > 11) {
    numbersOnly = numbersOnly.slice(0, 11);
  }

  // Aplica a máscara
  e.target.value = applyPhoneMask(numbersOnly);
});

// Permite apenas números, backspace, delete e setas
document.getElementById("phone").addEventListener("keydown", function (e) {
  // Permite: backspace, delete, tab, escape, enter
  if (
    [46, 8, 9, 27, 13].indexOf(e.keyCode) !== -1 ||
    // Permite: Ctrl+A, Ctrl+C, Ctrl+V, Ctrl+X
    (e.keyCode === 65 && e.ctrlKey === true) ||
    (e.keyCode === 67 && e.ctrlKey === true) ||
    (e.keyCode === 86 && e.ctrlKey === true) ||
    (e.keyCode === 88 && e.ctrlKey === true) ||
    // Permite: home, end, left, right
    (e.keyCode >= 35 && e.keyCode <= 39)
  ) {
    return;
  }
  // Garante que é um número
  if (
    (e.shiftKey || e.keyCode < 48 || e.keyCode > 57) &&
    (e.keyCode < 96 || e.keyCode > 105)
  ) {
    e.preventDefault();
  }
});
