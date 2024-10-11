package util

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func BlockIP(ip string, collection *mongo.Collection) error {
	_, err := collection.InsertOne(context.TODO(), bson.M{
		"blocked_ip": ip,
	})
	if err != nil {
		log.Printf("Erro ao bloquear o IP no MongoDB: %v", err)
		return err
	}

	log.Printf("IP %s bloqueado com sucesso", ip)
	return nil
}

// Função para verificar se o IP está bloqueado no MongoDB
func IsIPBlocked(ip string, blockedIPsCollection *mongo.Collection) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result bson.M
	err := blockedIPsCollection.FindOne(ctx, bson.M{"ip": ip}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false // IP não encontrado, não está bloqueado
		}
		log.Println("Erro ao verificar IP no MongoDB:", err)
		return false
	}
	return true // IP encontrado, está bloqueado
}

// Função para formatar e extrair o IP do endereço remoto (removendo a porta)
func GetClientIP(remoteAddr string) string {
	// Se o formato for algo como "192.168.0.1:12345", pegar só o IP
	if colonIndex := strings.LastIndex(remoteAddr, ":"); colonIndex != -1 {
		return remoteAddr[:colonIndex]
	}
	return remoteAddr
}

// Função para validar se a query ou payload bate com a regex proibida
func IsPayloadInvalid(payload string) bool {
	// Regex para validar conteúdo proibido
	regex := `<(?:a|abbr|acronym|address|applet|area|audioscope|b|base|basefront|bdo|bgsound|big|blackface|blink|blockquote|body|bq|br|button|caption|center|cite|code|col|colgroup|comment|dd|del|dfn|dir|div|dl|dt|em|embed|fieldset|fn|font|form|frame|frameset|h1|head|hr|html|i|iframe|ilayer|img|input|ins|isindex|kdb|keygen|label|layer|legend|li|limittext|link|listing|map|marquee|menu|meta|multicol|nobr|noembed|noframes|noscript|nosmartquotes|object|ol|optgroup|option|p|param|plaintext|pre|q|rt|ruby|s|samp|script|select|server|shadow|sidebar|small|spacer|span|strike|strong|style|sub|sup|table|tbody|td|textarea|tfoot|th|thead|title|tr|tt|u|ul|var|wbr|xml|xmp)\W`
	matched, _ := regexp.MatchString(regex, payload)
	return matched
}
