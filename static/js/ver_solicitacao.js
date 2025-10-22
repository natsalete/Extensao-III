// Initialize Lucide icons
document.addEventListener('DOMContentLoaded', function() {
    if (typeof lucide !== 'undefined') {
        lucide.createIcons();
    }
});

// Função para cancelar solicitação
function cancelarSolicitacao(id) {
    const form = document.getElementById("cancelForm");
    if (form) {
        form.action = "/solicitacao/" + id + "/cancelar";

        if (typeof bootstrap !== 'undefined') {
            const modalElement = document.getElementById("cancelModal");
            if (modalElement) {
                const modal = new bootstrap.Modal(modalElement);
                modal.show();
                
                // Reinicializar ícones quando o modal aparecer
                modalElement.addEventListener('shown.bs.modal', function () {
                    if (typeof lucide !== 'undefined') {
                        lucide.createIcons();
                    }
                }, { once: true });
            }
        }
    }
}

// Smooth scroll para seções
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        const href = this.getAttribute('href');
        if (href !== '#' && href.length > 1) {
            e.preventDefault();
            const target = document.querySelector(href);
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        }
    });
});

// Print functionality
const printBtn = document.querySelector('[onclick="window.print()"]');
if (printBtn) {
    printBtn.addEventListener('click', function(e) {
        e.preventDefault();
        window.print();
    });
}