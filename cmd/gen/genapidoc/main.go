package genapidoc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mitchellh/cli"
	"github.com/yusufsyaifudin/ngendika/extd"
	"github.com/yusufsyaifudin/ngendika/pkg/validator"
	"github.com/yusufsyaifudin/openapidoc/schema"
	"github.com/yusufsyaifudin/openapidoc/utils"
	"github.com/yusufsyaifudin/ylog"
	"log"
	"os"
	"path"
)

type ApiDocCfg struct {
}

type ApiDoc struct {
	Config ApiDocCfg
}

var _ cli.Command = (*ApiDoc)(nil)

func NewApiDocCmd(cfg ApiDocCfg) (*ApiDoc, error) {
	err := validator.Validate(cfg)
	if err != nil {
		err = fmt.Errorf("genapidocs: validation error: %w", err)
		return nil, err
	}

	return &ApiDoc{Config: cfg}, nil
}

func (a *ApiDoc) Help() string {
	return "generate apidoc json to be served on server"
}

func (a *ApiDoc) Synopsis() string {
	return "generate apidoc json to be served on server"
}

// Run .
// all responses must follow: respbuilder.RespStructure{}
func (a *ApiDoc) Run(args []string) int {
	ctx := context.Background()

	info := &openapi3.Info{
		Title:          "Ngendika",
		Description:    "",
		TermsOfService: "",
		Contact: &openapi3.Contact{
			Name: "Yusuf",
		},
		License: nil,
	}

	servers := openapi3.Servers{
		{
			URL:         "http://localhost:1234/",
			Description: "Localhost",
		},
	}

	components := openapi3.Components{
		ExtensionProps:  openapi3.ExtensionProps{},
		Schemas:         map[string]*openapi3.SchemaRef{},
		Parameters:      map[string]*openapi3.ParameterRef{},
		Headers:         map[string]*openapi3.HeaderRef{},
		RequestBodies:   map[string]*openapi3.RequestBodyRef{},
		Responses:       map[string]*openapi3.ResponseRef{},
		SecuritySchemes: map[string]*openapi3.SecuritySchemeRef{},
		Examples:        map[string]*openapi3.ExampleRef{},
		Links:           map[string]*openapi3.LinkRef{},
		Callbacks:       map[string]*openapi3.CallbackRef{},
	}
	paths := make(map[string]*openapi3.PathItem)

	// ** register default backends
	err := extd.RegisterDefaultBackends(ctx)
	if err != nil {
		ylog.Error(ctx, "register default backend failed", ylog.KV("error", err))
		return 1
	}

	// ** Register all routes here
	AppCreate(ctx, components, paths)
	AppCreateOrReplace(ctx, components, paths)
	AppDelete(ctx, components, paths)
	AppGetList(ctx, components, paths)
	AppGetOne(ctx, components, paths)
	PnpCreate(ctx, components, paths)

	doc := &openapi3.T{
		OpenAPI:    "3.0.0",
		Components: components,
		Info:       info,
		Servers:    servers,
		Paths:      paths,
	}

	j, err := doc.MarshalJSON()
	if err != nil {
		err = fmt.Errorf("cannot marshal openapi3 doc: %w", err)
		log.Println(err)
		return 1
	}

	var i interface{}
	err = json.Unmarshal(j, &i)
	if err != nil {
		err = fmt.Errorf("cannot unmarshal openapi3 doc: %w", err)
		log.Println(err)
		return 1
	}

	y, err := utils.YamlMarshalIndent(i)
	if err != nil {
		err = fmt.Errorf("cannot marshal YAML openapi3 doc: %w", err)
		log.Println(err)
		return 1
	}

	currDir, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("error get current dir: %w", err)
		log.Println(err)
		return 1
	}

	err = WriteFile(j, currDir+"/assets/swaggerui/swagger.json")
	if err != nil {
		log.Println(err)
		return 1
	}

	err = WriteFile(y, currDir+"/assets/swaggerui/swagger.yaml")
	if err != nil {
		log.Println(err)
		return 1
	}

	err = WriteFile(j, currDir+"/website/static/swagger.json")
	if err != nil {
		log.Println(err)
		return 1
	}

	err = WriteFile(y, currDir+"/website/static/swagger.yaml")
	if err != nil {
		log.Println(err)
		return 1
	}

	return 0
}

func WriteFile(content []byte, fileName string) (err error) {
	dir := path.Dir(fileName)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("cannot create directory %s: %w", dir, err)
		return
	}

	_, err = os.Stat(fileName)
	if errors.Is(err, os.ErrNotExist) {
		// create if not exist
		_, err = os.Create(fileName)
		if err != nil {
			err = fmt.Errorf("cannot create non-existed file %s: %w", fileName, err)
			return
		}

		err = nil
	}

	if err != nil {
		err = fmt.Errorf("cannot create file %s: %w", fileName, err)
		return
	}

	// Overwrite the file here.
	// Doesn't need to close because it uses in memory worktree.
	file, err := os.OpenFile(fileName, os.O_RDWR, os.ModePerm)
	defer func() {
		if file == nil {
			return
		}

		if _err := file.Close(); _err != nil {
			err = fmt.Errorf("cannot close file: %s: %w", fileName, _err)
			return
		}
	}()

	if err != nil {
		err = fmt.Errorf("cannot open file %s: %w", fileName, err)
		return
	}

	err = file.Truncate(0)
	if err != nil {
		err = fmt.Errorf("cannot truncate file %s: %w", fileName, err)
		return
	}

	_, err = file.Write(content)
	if err != nil {
		err = fmt.Errorf("cannot overwrite file %s: %w", fileName, err)
		return
	}

	return
}

func MustNewSchemaGenerator(ctx context.Context, prefix string, value interface{}) schema.GenerateOut {
	g, err := schema.NewGenerator(schema.WithLog(os.Stdout), schema.WithSchemaPrefix(prefix))
	if err != nil {
		panic(err)
	}

	out, err := g.Generate(ctx, value)
	if err != nil {
		panic(err)
	}

	return out
}
