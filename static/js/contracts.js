// ============================================
// Contracts JavaScript - Versão Melhorada
// ============================================

console.log('✅ Contracts JS carregado com sucesso!');

// Funções auxiliares para contratos
const ContractsHelper = {
    
    // Formatar valor em reais
    formatCurrency: function(value) {
        return new Intl.NumberFormat('pt-BR', {
            style: 'currency',
            currency: 'BRL'
        }).format(value);
    },

    // Formatar data
    formatDate: function(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('pt-BR');
    },

    // Confirmar ação
    confirmAction: function(message) {
        return confirm(message);
    },

    // Copiar texto
    copyToClipboard: function(text) {
        navigator.clipboard.writeText(text).then(() => {
            alert('Copiado para a área de transferência!');
        }).catch(err => {
            console.error('Erro ao copiar:', err);
        });
    }
};

// ============================================
// IMPRESSÃO DE CONTRATOS EM PDF
// ============================================

/**
 * Imprime o contrato atual
 * Garante que as assinaturas sejam renderizadas corretamente
 */
function printContract() {
    console.log('🖨️ Iniciando impressão do contrato...');
    
    // Verificar se as imagens de assinatura carregaram
    const signatureImages = document.querySelectorAll('.signature-img');
    let allLoaded = true;
    
    signatureImages.forEach((img, index) => {
        if (!img.complete || img.naturalHeight === 0) {
            console.warn(`⚠️ Assinatura ${index + 1} não carregada completamente`);
            allLoaded = false;
        } else {
            console.log(`✅ Assinatura ${index + 1} carregada: ${img.naturalWidth}x${img.naturalHeight}`);
        }
    });
    
    if (!allLoaded) {
        console.log('⏳ Aguardando carregamento das assinaturas...');
        
        // Aguardar 500ms para garantir que as imagens carreguem
        setTimeout(() => {
            console.log('🖨️ Executando impressão...');
            window.print();
        }, 500);
    } else {
        console.log('🖨️ Executando impressão...');
        window.print();
    }
}

/**
 * Exporta contrato para PDF (alternativa usando html2pdf)
 * Requer: <script src="https://cdnjs.cloudflare.com/ajax/libs/html2pdf.js/0.10.1/html2pdf.bundle.min.js"></script>
 */
function exportContractToPDF() {
    console.log('📄 Exportando contrato para PDF...');
    
    const contractElement = document.getElementById('contract-content');
    const signaturesElement = document.getElementById('signatures-section');
    
    if (!contractElement) {
        alert('Erro: Conteúdo do contrato não encontrado');
        return;
    }
    
    // Verificar se html2pdf está disponível
    if (typeof html2pdf === 'undefined') {
        console.warn('html2pdf não disponível. Usando impressão padrão...');
        printContract();
        return;
    }
    
    // Criar container temporário com conteúdo completo
    const container = document.createElement('div');
    container.style.padding = '20px';
    container.appendChild(contractElement.cloneNode(true));
    
    if (signaturesElement) {
        container.appendChild(signaturesElement.cloneNode(true));
    }
    
    // Configurações do PDF
    const opt = {
        margin: 1,
        filename: `contrato-${Date.now()}.pdf`,
        image: { type: 'jpeg', quality: 0.98 },
        html2canvas: { 
            scale: 2,
            useCORS: true,
            logging: false
        },
        jsPDF: { 
            unit: 'cm', 
            format: 'a4', 
            orientation: 'portrait' 
        }
    };
    
    // Gerar PDF
    html2pdf().set(opt).from(container).save().then(() => {
        console.log('✅ PDF gerado com sucesso!');
    }).catch(err => {
        console.error('❌ Erro ao gerar PDF:', err);
        alert('Erro ao gerar PDF. Tentando impressão padrão...');
        printContract();
    });
}

// ============================================
// DIAGNÓSTICO DE ASSINATURAS
// ============================================

/**
 * Verifica e diagnostica problemas com assinaturas
 */
function diagnoseSignatures() {
    console.log('🔍 Diagnóstico de Assinaturas:');
    console.log('================================');
    
    const signatureImages = document.querySelectorAll('.signature-img');
    
    if (signatureImages.length === 0) {
        console.warn('⚠️ Nenhuma imagem de assinatura encontrada no DOM');
        return;
    }
    
    signatureImages.forEach((img, index) => {
        console.log(`\n📝 Assinatura #${index + 1}:`);
        console.log('  - Elemento:', img);
        console.log('  - src presente:', !!img.src);
        console.log('  - src length:', img.src?.length || 0);
        console.log('  - src preview:', img.src?.substring(0, 80) + '...');
        console.log('  - complete:', img.complete);
        console.log('  - naturalWidth:', img.naturalWidth);
        console.log('  - naturalHeight:', img.naturalHeight);
        console.log('  - display:', window.getComputedStyle(img).display);
        console.log('  - visibility:', window.getComputedStyle(img).visibility);
        
        // Verificar erros
        if (img.naturalWidth === 0 && img.complete) {
            console.error('  ❌ ERRO: Imagem não pôde ser carregada (src inválido)');
        } else if (!img.complete) {
            console.warn('  ⏳ Aguardando carregamento...');
        } else {
            console.log('  ✅ Imagem OK');
        }
    });
    
    console.log('================================');
}

/**
 * Força recarregamento de assinaturas com erro
 */
function reloadFailedSignatures() {
    console.log('🔄 Recarregando assinaturas com erro...');
    
    const signatureImages = document.querySelectorAll('.signature-img');
    let reloadCount = 0;
    
    signatureImages.forEach((img) => {
        if (img.naturalWidth === 0 && img.complete) {
            console.log('🔄 Recarregando:', img.alt);
            const src = img.src;
            img.src = '';
            setTimeout(() => {
                img.src = src;
            }, 100);
            reloadCount++;
        }
    });
    
    if (reloadCount === 0) {
        console.log('✅ Nenhuma assinatura precisa ser recarregada');
    } else {
        console.log(`🔄 ${reloadCount} assinatura(s) recarregada(s)`);
    }
}

// ============================================
// INICIALIZAÇÃO
// ============================================

// Fazer funções disponíveis globalmente
window.ContractsHelper = ContractsHelper;
window.printContract = printContract;
window.exportContractToPDF = exportContractToPDF;
window.diagnoseSignatures = diagnoseSignatures;
window.reloadFailedSignatures = reloadFailedSignatures;

// Inicialização quando DOM estiver pronto
document.addEventListener('DOMContentLoaded', function() {
    console.log('📄 Página de contratos inicializada');
    
    // Adicionar tooltips do Bootstrap se disponível
    if (typeof bootstrap !== 'undefined') {
        const tooltipTriggerList = [].slice.call(
            document.querySelectorAll('[data-bs-toggle="tooltip"]')
        );
        tooltipTriggerList.map(function (tooltipTriggerEl) {
            return new bootstrap.Tooltip(tooltipTriggerEl);
        });
    }
    
    // Monitorar carregamento de assinaturas
    const signatureImages = document.querySelectorAll('.signature-img');
    
    signatureImages.forEach((img, index) => {
        // Log quando carregar com sucesso
        img.addEventListener('load', function() {
            console.log(`✅ Assinatura ${index + 1} carregada com sucesso (${this.naturalWidth}x${this.naturalHeight})`);
        });
        
        // Log quando houver erro
        img.addEventListener('error', function() {
            console.error(`❌ Erro ao carregar assinatura ${index + 1}`);
            console.error('   src:', this.src.substring(0, 100) + '...');
            
            // Mostrar mensagem de erro visível
            const errorDiv = this.nextElementSibling;
            if (errorDiv) {
                errorDiv.style.display = 'block';
            }
            this.style.display = 'none';
        });
    });
    
    // Diagnóstico automático após 2 segundos
    setTimeout(() => {
        diagnoseSignatures();
    }, 2000);
    
    // Adicionar event listeners para botões de impressão
    const printButtons = document.querySelectorAll('[onclick*="printContract"]');
    printButtons.forEach(btn => {
        console.log('🖨️ Botão de impressão detectado:', btn);
    });
});

// Listener para antes de imprimir
window.addEventListener('beforeprint', function() {
    console.log('🖨️ Preparando para impressão...');
    
    // Garantir que elementos no-print estejam escondidos
    document.querySelectorAll('.no-print').forEach(el => {
        el.style.display = 'none';
    });
});

// Listener para depois de imprimir
window.addEventListener('afterprint', function() {
    console.log('✅ Impressão concluída');
    
    // Restaurar elementos no-print
    document.querySelectorAll('.no-print').forEach(el => {
        el.style.display = '';
    });
});