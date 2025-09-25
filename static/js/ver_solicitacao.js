// Initialize Lucide icons
lucide.createIcons();

// Função para cancelar solicitação
function cancelarSolicitacao(id) {
  const form = document.getElementById("cancelForm");
  form.action = "/solicitacao/" + id + "/cancelar";

  const modal = new bootstrap.Modal(document.getElementById("cancelModal"));
  modal.show();
}
