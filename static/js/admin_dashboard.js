// ========================================
// ADMIN DASHBOARD - JavaScript
// ========================================

// Limpar filtros
function clearFilters() {
  window.location.href = '/dashboard/admin';
}

// ✅ CORRIGIDO - Abrir modal de alteração de status
function openStatusModal(requestId, currentStatusId, clientName) {
  // Verificar se os elementos existem antes de usar
  const modalRequestIdInput = document.querySelector('#statusModal input[name="request_id"]');
  const modalClientNameSpan = document.getElementById('modalClientName');
  const modalStatusSelect = document.querySelector('#statusModal select[name="status_id"]');
  
  if (modalRequestIdInput) {
    modalRequestIdInput.value = requestId;
  }
  
  if (modalClientNameSpan) {
    modalClientNameSpan.textContent = clientName;
  }
  
  if (modalStatusSelect) {
    modalStatusSelect.value = currentStatusId;
  }
  
  // Abrir o modal
  const modalElement = document.getElementById('statusModal');
  if (modalElement) {
    const modal = new bootstrap.Modal(modalElement);
    modal.show();
  } else {
    console.error('Modal statusModal não encontrado!');
  }
}

// Deletar solicitação
function deletarSolicitacao(requestId) {
  const deleteForm = document.getElementById('deleteForm');
  if (deleteForm) {
    deleteForm.action = `/admin/solicitacao/${requestId}/deletar`;
    
    const modalElement = document.getElementById('deleteModal');
    if (modalElement) {
      const modal = new bootstrap.Modal(modalElement);
      modal.show();
    }
  }
}

// Auto-hide success messages
document.addEventListener('DOMContentLoaded', function() {
  // Auto-hide alerts after 5 seconds
  const alerts = document.querySelectorAll('.alert');
  alerts.forEach(alert => {
    setTimeout(() => {
      const bsAlert = new bootstrap.Alert(alert);
      bsAlert.close();
    }, 5000);
  });

  // Highlight row on hover
  const rows = document.querySelectorAll('tbody tr');
  rows.forEach(row => {
    row.addEventListener('mouseenter', function() {
      this.style.backgroundColor = '#f8f9fa';
    });
    row.addEventListener('mouseleave', function() {
      this.style.backgroundColor = '';
    });
  });
});

// Export table to CSV
function exportToCSV() {
  const table = document.querySelector('table');
  if (!table) {
    alert('Nenhuma tabela encontrada para exportar');
    return;
  }

  let csv = [];
  const rows = table.querySelectorAll('tr');
  
  rows.forEach(row => {
    const cols = row.querySelectorAll('td, th');
    const csvRow = [];
    cols.forEach(col => {
      let text = col.innerText.trim();
      text = text.replace(/"/g, '""');
      csvRow.push(`"${text}"`);
    });
    csv.push(csvRow.join(','));
  });
  
  const csvContent = csv.join('\n');
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `solicitacoes_${new Date().toISOString().split('T')[0]}.csv`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  window.URL.revokeObjectURL(url);
}

// Print table
function printTable() {
  window.print();
}

// Confirm action
function confirmAction(message, callback) {
  if (confirm(message)) {
    callback();
  }
}

// Show loading overlay
function showLoading() {
  const overlay = document.createElement('div');
  overlay.className = 'loading-overlay';
  overlay.innerHTML = `
    <div class="spinner-border text-primary" role="status">
      <span class="visually-hidden">Carregando...</span>
    </div>
  `;
  document.body.appendChild(overlay);
}

// Hide loading overlay
function hideLoading() {
  const overlay = document.querySelector('.loading-overlay');
  if (overlay) {
    overlay.remove();
  }
}

// Format date to Brazilian format
function formatDate(dateString) {
  const date = new Date(dateString);
  return date.toLocaleDateString('pt-BR');
}

// Format time
function formatTime(timeString) {
  return timeString.slice(0, 5);
}

// Copy to clipboard
function copyToClipboard(text) {
  navigator.clipboard.writeText(text).then(() => {
    const toast = document.createElement('div');
    toast.className = 'toast show position-fixed bottom-0 end-0 m-3';
    toast.innerHTML = `
      <div class="toast-body bg-success text-white">
        <i class="bi bi-check-circle me-2"></i>
        Copiado para a área de transferência!
      </div>
    `;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
  }).catch(err => {
    console.error('Erro ao copiar:', err);
  });
}

// Initialize tooltips
document.addEventListener('DOMContentLoaded', function() {
  const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
  tooltipTriggerList.map(function (tooltipTriggerEl) {
    return new bootstrap.Tooltip(tooltipTriggerEl);
  });
});

// Search with debounce
let searchTimeout;
function debounceSearch(input, delay = 500) {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    input.form.submit();
  }, delay);
}

// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
  // Ctrl/Cmd + K - Focus search
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault();
    const searchInput = document.querySelector('input[name="search"]');
    if (searchInput) {
      searchInput.focus();
    }
  }
  
  // Ctrl/Cmd + P - Print
  if ((e.ctrlKey || e.metaKey) && e.key === 'p') {
    e.preventDefault();
    printTable();
  }
  
  // Esc - Close modals
  if (e.key === 'Escape') {
    const modals = document.querySelectorAll('.modal.show');
    modals.forEach(modal => {
      const bsModal = bootstrap.Modal.getInstance(modal);
      if (bsModal) {
        bsModal.hide();
      }
    });
  }
});

// Update stats in real-time
function updateStats() {
  console.log('Stats updated');
}

// Refresh page data
function refreshData() {
  showLoading();
  window.location.reload();
}

// Sort table
function sortTable(columnIndex) {
  const table = document.querySelector('table');
  const tbody = table.querySelector('tbody');
  const rows = Array.from(tbody.querySelectorAll('tr'));
  
  rows.sort((a, b) => {
    const aText = a.cells[columnIndex].innerText;
    const bText = b.cells[columnIndex].innerText;
    return aText.localeCompare(bText);
  });
  
  rows.forEach(row => tbody.appendChild(row));
}