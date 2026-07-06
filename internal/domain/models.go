package domain

import "sync"

var (
	Exit = make(chan bool)
	// exitOnce ensures the shutdown channel is closed only once.
	exitOnce sync.Once
)

func CloseExit() {
	exitOnce.Do(func() {
		close(Exit)
	})
}

type UserState struct {
	Locale            string
	WaitingForTramite bool
}

type TramiteResponse struct {
	Data    TramiteData `json:"data"`
	Mensaje string      `json:"mensaje"`
	Codigo  int         `json:"codigo"`
}

type TramiteData struct {
	OficinaRemitente            OficinaRemitente `json:"oficina_remitente"`
	EnOperativo                 EnOperativo      `json:"en_operativo"`
	FechaToma                   string           `json:"fecha_toma"`
	IDUltimoEstado              string           `json:"id_ultimo_estado"`
	DescripcionUltimoEstado     string           `json:"descripcion_ultimo_estado"`
	IDOficinaOrigen             string           `json:"id_oficina_origen"`
	FechaUltimoEstado           string           `json:"fecha_ultimo_estado"`
	TipoTramite                 string           `json:"tipo_tramite"`
	Accion                      string           `json:"accion"`
	Motivo                      string           `json:"motivo"`
	IDAnteultimoEstado          string           `json:"id_anteultimo_estado"`
	DescripcionAnteultimoEstado string           `json:"descripcion_anteultimo_estado"`
	FechaAnteultimoEstado       string           `json:"fecha_anteultimo_estado"`
	TipoRetiro                  string           `json:"tipo_retiro"`
	Correo                      string           `json:"correo"`
	Direccion                   string           `json:"direccion"`
	IDTramite                   string           `json:"id_tramite"`
	DescripcionTramite          string           `json:"descripcion_tramite"`
	ClaseTramite                string           `json:"clase_tramite"`
	TipoDNI                     string           `json:"tipo_dni"`
	TramitesUI                  []TramiteUI      `json:"tramitesUI"`
	IDCorreo                    int              `json:"id_correo"`
	Operativo                   int              `json:"operativo"`
}

type OficinaRemitente struct {
	Codigo       string `json:"codigo"`
	Descripcion  string `json:"descripcion"`
	Domicilio    string `json:"domicilio"`
	CodigoPostal string `json:"codigo_postal"`
	Provincia    string `json:"provincia"`
}

type EnOperativo struct {
	Operativo int `json:"operativo"`
}

type TramiteUI struct {
	Historico       []HistoricoItem `json:"historico"`
	Descripcion     string          `json:"descripcion"`
	EstadoLabel     string          `json:"estado_label"`
	Icon            string          `json:"icon"`
	ClassPanelColor string          `json:"class_panel_color"`
	TextoFront      string          `json:"texto_front"`
	TextoCuerpo     string          `json:"texto_cuerpo"`
	SVG             string          `json:"svg"`
	TipoTramite     string          `json:"tipo_tramite"`
	TipoDNI         string          `json:"tipo_dni"`
	EstadosCHUTRO   []int           `json:"estados_CHUTRO"`
}

type HistoricoItem struct {
	CodPlanta    string `json:"codPlanta"`
	Estado       string `json:"estado"`
	Evento       string `json:"evento"`
	Fecha        string `json:"fecha"`
	FechaAux     string `json:"fechaAux"`
	Firma        string `json:"firma"`
	Planta       string `json:"planta"`
	FechaCambio  string `json:"FechaCambio"`
	EstadoMotivo string `json:"EstadoMotivo"`
}
