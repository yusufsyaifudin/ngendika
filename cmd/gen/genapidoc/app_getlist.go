package genapidoc

import (
	"context"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/yusufsyaifudin/ngendika/pkg/respbuilder"
	"github.com/yusufsyaifudin/ngendika/transport/restapi/handlerapp"
	httptyped2 "github.com/yusufsyaifudin/ngendika/transport/restapi/httptyped"
	"math"
	"net/http"
	"time"
)

// AppGetList
// GET /api/v1/apps
func AppGetList(ctx context.Context, components openapi3.Components, path map[string]*openapi3.PathItem) {
	const scopedSchemaName = "AppGetList"
	const routeName = "Get ListByProvider of Applications"
	const pathRoute = "/api/v1/apps"

	// --- Response schema
	respStruct := handlerapp.ListAppsResp{
		Total: 1,
		Limit: 100,
		Items: []httptyped2.AppEntity{
			{
				ID:        123,
				ClientID:  "myapp",
				Name:      "My App",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	// generate response and add to components
	resp := respbuilder.Success(ctx, respStruct)
	outResp := MustNewSchemaGenerator(ctx, scopedSchemaName+".Resp201.", resp)
	for s, ref := range outResp.Schemas {
		components.Schemas[s] = ref
	}

	// --- params
	paramLimit := openapi3.NewQueryParameter("limit").WithDescription("Limit of list returned")
	paramLimit.Schema = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "number"}}
	paramLimit.Example = 100
	paramLimit.Required = false

	paramBeforeID := openapi3.NewQueryParameter("max_id").WithDescription("Maximum of app id to be fetched")
	paramBeforeID.Schema = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "number"}}
	paramBeforeID.Example = fmt.Sprintf("%d", math.MaxInt64)
	paramBeforeID.Required = false

	paramAfterID := openapi3.NewQueryParameter("min_id").WithDescription("Minimum of app id to be fetched")
	paramAfterID.Schema = &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "number"}}
	paramAfterID.Example = 0
	paramAfterID.Required = false

	// --- final spec
	op := openapi3.NewOperation()
	op.Tags = []string{"Application"}
	op.Summary = routeName
	op.OperationID = scopedSchemaName
	op.AddParameter(paramLimit)
	op.AddParameter(paramBeforeID)
	op.AddParameter(paramAfterID)

	op.AddResponse(http.StatusOK, openapi3.NewResponse().WithJSONSchemaRef(
		&openapi3.SchemaRef{
			Ref: fmt.Sprintf("#/components/schemas/%s", outResp.ParentSchemaName),
		},
	).WithDescription("desc"))

	_, exist := path[pathRoute]
	if !exist {
		path[pathRoute] = &openapi3.PathItem{}
	}

	path[pathRoute].Get = op
}
