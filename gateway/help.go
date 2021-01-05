package gateway

func RenderTable(args [][]string) string {
	back := ""
	for _, x := range args {
		back += "<tr>"
		for _, y := range x {
			back += "<td>" + y + "</td>"
		}
		back += "</tr>"
	}
	return back
}
