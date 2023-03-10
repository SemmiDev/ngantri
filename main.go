package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"ngantri/pawnshop"
	"ngantri/queue"
	"ngantri/token"
	"ngantri/utils"
	"path"
)

//go:embed views/*
var Resources embed.FS

//go:embed public
var StaticFiles embed.FS

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	symmetricKey := utils.GetEnv("SYMMETRIC_KEY", "76440225709465899502497877279549")
	tokenMaker, err := token.NewPasetoMaker(symmetricKey)
	if err != nil {
		log.Fatalf("failed to create token maker: %v", err)
	}

	db, dbErr := utils.ConnectDB()
	if dbErr != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	handlers := &handlers{
		pawnshopDataStore: pawnshop.NewMySqlDataStore(db),
		queueDataStore:    queue.NewMySqlDataStore(db),
		tokenMaker:        tokenMaker,
	}

	r.Get("/public/*", http.StripPrefix("/public", fsHandler()).ServeHTTP)

	r.Get("/", handlers.homePageHandler)
	r.Get("/pawnshops/form", handlers.pawnShopFormHandler)
	r.Post("/pawnshops/form/process", handlers.pawnShopFormProcessHandler)
	r.Get("/queues/request", handlers.queueRequestHandler)

	r.Get("/queues/status/{id}", handlers.updateQueueStatusNumberHandler)
	r.Get("/queues/today", handlers.getAllQueueByCurrentDateHandler)
	r.Get("/queues", handlers.getAllQueueByCurrentDateHandler)

	serverPort := utils.GetEnv("PORT", "8080")
	fmt.Println("Server started on port", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, r))
}

type handlers struct {
	pawnshopDataStore *pawnshop.MySqlDataStore
	queueDataStore    *queue.MySqlDataStore
	tokenMaker        token.Maker
}

func fsHandler() http.Handler {
	sub, err := fs.Sub(StaticFiles, "public")
	if err != nil {
		log.Fatalf("failed to open static files: %v", err)
	}
	return http.FileServer(http.FS(sub))
}

func (h *handlers) homePageHandler(w http.ResponseWriter, _ *http.Request) {
	homePage := template.Must(template.ParseFS(Resources, path.Join("views", "index.html")))
	homePage.Execute(w, nil)
}

func (h *handlers) pawnShopFormHandler(w http.ResponseWriter, _ *http.Request) {
	pawnshopFormPage := template.Must(template.ParseFS(Resources, path.Join("views", "pawnshop-form.html")))
	pawnshopFormPage.Execute(w, nil)
}

func (h *handlers) queueRequestHandler(w http.ResponseWriter, _ *http.Request) {
	queueRequestResult, err := h.queueDataStore.GetAvailableQueue()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	queueRequest := template.Must(template.ParseFS(Resources, path.Join("views", "queue-request.html")))
	queueRequest.Execute(w, queueRequestResult)
}

func (h *handlers) updateQueueStatusNumberHandler(w http.ResponseWriter, r *http.Request) {
	queueID := chi.URLParam(r, "id")

	err := h.queueDataStore.ChangeQueueStatus(queueID, queue.Served)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handlers) getAllQueueByCurrentDateHandler(w http.ResponseWriter, r *http.Request) {
	queueList, err := h.queueDataStore.GetAllQueueByCurrentDate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(queueList)
}

func (h *handlers) pawnShopFormProcessHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	namaLengkap := r.FormValue("nama-lengkap")
	asalBarangJaminan := r.FormValue("asal-barang-jaminan")
	statusTransaksi := r.FormValue("status-transaksi")

	tujuanTransaksi := r.FormValue("tujuan-transaksi")
	tujuanTransaksiLainLain := r.FormValue("tujuan-transaksi-lain-lain")
	if tujuanTransaksi == "Lain-lain" || tujuanTransaksiLainLain != "" {
		tujuanTransaksi = tujuanTransaksiLainLain
	}

	caraPembayaran := r.FormValue("cara-pembayaran")

	fiturYangDiinginkan := r.FormValue("fitur-yang-diinginkan")
	lamaFleksi := r.FormValue("lama-fleksi")
	if lamaFleksi != "" {
		fiturYangDiinginkan += " " + lamaFleksi
	}

	pengambilanUangNamaBank := r.FormValue("pengambilan-uang-nama-bank")
	pengambilanUangNoRek := r.FormValue("pengambilan-uang-no-rek")
	pengambilanUangAn := r.FormValue("pengambilan-uang-an")
	pengambilanUang := fmt.Sprintf("Nomor Rekening: %s, Atas Nama: %s, Bank: %s", pengambilanUangNoRek, pengambilanUangAn, pengambilanUangNamaBank)

	kelebihanLelang := r.FormValue("kelebihan-lelang")
	topUpTabunganEmasNamaBank := r.FormValue("top-up-tabungan-emas-nama-bank")
	topUpTabunganEmasNoRek := r.FormValue("top-up-tabungan-emas-no-rek")
	topUpTabunganEmasAn := r.FormValue("top-up-tabungan-emas-an")
	kelebihanLelang += fmt.Sprintf(" Top Up Tabungan Emas: Nomor Rekening: %s, Atas Nama: %s, Bank: %s", topUpTabunganEmasNoRek, topUpTabunganEmasAn, topUpTabunganEmasNamaBank)

	besarPinjaman := r.FormValue("besar-pinjaman")
	besarPinjamanPermintaan := r.FormValue("besar-pinjaman-permintaan")
	if besarPinjaman == "Permintaan" {
		besarPinjaman = "Permintaaan sebesar " + besarPinjamanPermintaan
	}

	barangJaminan := r.FormValue("barang-jaminan")
	alamat := r.FormValue("alamat")
	noHp := r.FormValue("nomor-hp")

	pawnshopData := pawnshop.Pawnshop{
		NamaLengkap:         namaLengkap,
		AsalBarangJaminan:   &asalBarangJaminan,
		StatusTransaksi:     &statusTransaksi,
		TujuanTransaksi:     &tujuanTransaksi,
		CaraPembayaran:      caraPembayaran,
		FiturYangDiinginkan: &fiturYangDiinginkan,
		PengambilanUang:     pengambilanUang,
		KelebihanLelang:     &kelebihanLelang,
		BesarPinjaman:       &besarPinjaman,
		BarangJaminan:       barangJaminan,
		Alamat:              alamat,
		NoHP:                noHp,
	}

	err = h.pawnshopDataStore.AddPawnshop(pawnshopData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/queues/request", http.StatusSeeOther)
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}
