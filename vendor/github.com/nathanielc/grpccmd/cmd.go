package grpccmd

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var rootCmd = &cobra.Command{}

var addr = new(string)
var input = new(string)

func init() {
	rootCmd.PersistentFlags().StringVar(addr, "addr", "", "gRPC server address.")
	rootCmd.PersistentFlags().StringVar(input, "input", "", "JSON representation of the input data for the method.")
}

func SetCmdInfo(name, short, long string) {
	rootCmd.Use = fmt.Sprintf("%s [command]", name)
	rootCmd.Short = short
	rootCmd.Long = long
}

func RegisterServiceCmd(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func RunE(
	method,
	inT string,
	newClient func(*grpc.ClientConn) interface{},
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		conn, err := dial(*addr)
		if err != nil {
			return err
		}
		defer conn.Close()
		c := newClient(conn)
		cv := reflect.ValueOf(c)
		method := cv.MethodByName(method)
		if method.IsValid() {

			in := reflect.New(proto.MessageType(inT).Elem()).Interface()
			if len(*input) > 0 {
				if err := json.Unmarshal([]byte(*input), in); err != nil {
					return err
				}
			}

			result := method.Call([]reflect.Value{
				reflect.ValueOf(context.Background()),
				reflect.ValueOf(in),
			})
			if len(result) != 2 {
				panic("service methods should always return 2 values")
			}
			if !result[1].IsNil() {
				return result[1].Interface().(error)
			}
			out := result[0].Interface()
			data, err := json.MarshalIndent(out, "", "    ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		}

		return nil
	}
}

func dial(addr string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	return grpc.Dial(addr, opts...)
}
