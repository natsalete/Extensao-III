package controllers

import (
	"net/http"
	"time"

	"martins-pocos/models"
	"martins-pocos/utils"
)

// createServiceRequest - Processa criação de nova solicitação
func (c *ServiceController) createServiceRequest(w http.ResponseWriter, r *http.Request, userID int) {
	if err := r.ParseForm(); err != nil {
		utils.SendErrorResponse(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	service, err := c.parseServiceRequestForm(r, userID)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.ServiceModel.Create(service); err != nil {
		utils.SendErrorResponse(w, "Erro ao criar solicitação", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard/cliente?success=created", http.StatusFound)
}

// updateServiceRequest - Processa atualização de solicitação
func (c *ServiceController) updateServiceRequest(w http.ResponseWriter, r *http.Request, requestID, userID int) {
	if err := r.ParseForm(); err != nil {
		utils.SendErrorResponse(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	service, err := c.parseServiceRequestForm(r, userID)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	service.ID = requestID

	if err := c.ServiceModel.Update(service); err != nil {
		c.handleUpdateError(w, err, "Solicitação não pode ser editada")
		return
	}

	http.Redirect(w, r, "/dashboard/cliente?success=updated", http.StatusFound)
}

// parseServiceRequestForm - Parseia o formulário de solicitação
func (c *ServiceController) parseServiceRequestForm(r *http.Request, userID int) (*models.ServiceRequest, error) {
	serviceTypeCode := r.FormValue("service_type")
	serviceType, err := c.ServiceModel.GetServiceTypeByCode(serviceTypeCode)
	if err != nil {
		return nil, err
	}

	preferredDate, err := time.Parse("2006-01-02", r.FormValue("preferred_date"))
	if err != nil {
		return nil, err
	}

	preferredTime, err := time.Parse("15:04", r.FormValue("preferred_time"))
	if err != nil {
		return nil, err
	}

	return &models.ServiceRequest{
		UserID:        userID,
		FullName:      r.FormValue("full_name"),
		ServiceTypeID: serviceType.ID,
		Description:   r.FormValue("description"),
		CEP:           r.FormValue("cep"),
		Logradouro:    r.FormValue("logradouro"),
		Numero:        r.FormValue("numero"),
		Bairro:        r.FormValue("bairro"),
		Cidade:        r.FormValue("cidade"),
		Estado:        r.FormValue("estado"),
		PreferredDate: preferredDate,
		PreferredTime: preferredTime.Format("15:04"),
	}, nil
}