package lib

const (
	TPLuuid = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body>
    <p>New UUID4: {{.newuuid4}}</p>
    <br>
    <br>
    <br>
    <br>
    <br>
    <br>
    <br>
    <br>
    <br>
    <br>
    <p style="color:white;">id1: {{.uuid1}}</p>
    <p style="color:white;">id2: {{.uuid2}}</p>
    <p style="color:white;">id3: {{.uuid3}}</p>
    <p style="color:white;">id4: {{.uuid4}}</p>
    <p style="color:white;">id5: {{.uuid5}}</p>
</body>
</html>`

	TPLVpsinfo = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body>
    <div class="container">
        <table>
            <tr>
                <td width=200px>
                    <b>plan:</b>
                </td>
                <td>
                    {{.plan}}
                </td>
            </tr>
            <tr>
                <td>
                    <b>vm type:</b>
                </td>
                <td>
                    {{.vmtype}}
                </td>
            </tr>
            <tr>
                <td>
                    <b>os:</b>
                </td>
                <td>
                    {{.os}}
                </td>
            </tr>
            <tr>
                <td>
                    <b>hostname:</b>
                </td>
                <td>
                    {{.hostname}}
                </td>
            </tr>
            <tr>
                <td>
                    <b>location:</b>
                </td>
                <td>
                    {{.location}}
                </td>
            </tr>
            <tr>
                <td>
                    <b>datacenter:</b>
                </td>
                <td>
                    {{.datacenter}}
                </td>
            </tr>
            <tr>
                <td>
                    <b>monthly data:</b>
                </td>
                <td>
                    {{.plan_monthly_data}} GB
                </td>
            </tr>
            <tr>
                <td>
                    <b>used data:</b>
                </td>
                <td>
                    {{.data_counter}} GB
                </td>
            </tr>
            <tr>
                <td>
                    <b>support ipv6:</b>
                </td>
                <td>
                    {{.ivp6}}
                </td>
            </tr>
            <tr>
                <td>
                    <b>error:</b>
                </td>
                <td>
                    {{.error}}
                </td>
            </tr>
            <tr style="color:white;">
                <td>
                    <b>ip address:</b>
                </td>
                <td>
                    {{.ipv4}}
                </td>
            </tr>
        </table>
    </div>
</body>
</html>`
)
