// Initialize Lucide icons
lucide.createIcons();

// Count stats on page load
$(document).ready(function () {
  updateStats();
});

function updateStats() {
  let pending = 0,
    progress = 0,
    completed = 0;

  $(".status-badge").each(function () {
    if ($(this).hasClass("status-pendente")) pending++;
    else if ($(this).hasClass("status-em_andamento")) progress++;
    else if ($(this).hasClass("status-concluido")) completed++;
  });

  $("#pending-count").text(pending);
  $("#progress-count").text(progress);
  $("#completed-count").text(completed);
  $("#total-count").text(pending + progress + completed);
}

function updateStatus(requestId, newStatus) {
  $.ajax({
    url: "/update-status",
    method: "POST",
    data: {
      request_id: requestId,
      status: newStatus,
    },
    success: function (response) {
      // Update status badge
      const row = $(`tr[data-request-id="${requestId}"]`);
      const statusCell = row.find(".status-cell");

      let badgeClass = "";
      let statusText = "";

      switch (newStatus) {
        case "pendente":
          badgeClass = "status-pendente";
          statusText = "Pendente";
          break;
        case "em_andamento":
          badgeClass = "status-em_andamento";
          statusText = "Em Andamento";
          break;
        case "concluido":
          badgeClass = "status-concluido";
          statusText = "Concluído";
          break;
      }

      statusCell.html(
        `<span class="status-badge ${badgeClass}">${statusText}</span>`
      );

      // Update action buttons
      updateActionButtons(row, newStatus);

      // Update stats
      updateStats();

      // Show toast notification
      const toast = new bootstrap.Toast(document.getElementById("statusToast"));
      toast.show();
    },
    error: function () {
      alert("Erro ao atualizar status. Tente novamente.");
    },
  });
}

function updateActionButtons(row, currentStatus) {
  const actionCell = row.find("td:last-child .btn-group-vertical");
  actionCell.empty();

  if (currentStatus !== "pendente") {
    actionCell.append(`
                    <button class="btn btn-status btn-warning-custom" onclick="updateStatus(${row.data(
                      "request-id"
                    )}, 'pendente')">
                        Pendente
                    </button>
                `);
  }
  if (currentStatus !== "em_andamento") {
    actionCell.append(`
                    <button class="btn btn-status btn-info-custom" onclick="updateStatus(${row.data(
                      "request-id"
                    )}, 'em_andamento')">
                        Em Andamento
                    </button>
                `);
  }
  if (currentStatus !== "concluido") {
    actionCell.append(`
                    <button class="btn btn-status btn-success-custom" onclick="updateStatus(${row.data(
                      "request-id"
                    )}, 'concluido')">
                        Concluído
                    </button>
                `);
  }
}
