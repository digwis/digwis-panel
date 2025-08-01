// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.924
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func Notification() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<!-- 通知容器 --><div id=\"notificationContainer\" class=\"fixed top-4 right-4 z-50 space-y-2\"><!-- 通知会动态添加到这里 --></div><script>\n\t\t// 显示通知\n\t\tfunction showNotification(message, type = 'info', duration = 5000) {\n\t\t\tconst container = document.getElementById('notificationContainer');\n\t\t\tconst notificationId = 'notification-' + Date.now();\n\t\t\t\n\t\t\t// 确定通知样式\n\t\t\tlet bgColor, textColor, iconSvg;\n\t\t\tswitch(type) {\n\t\t\t\tcase 'success':\n\t\t\t\t\tbgColor = 'bg-green-500';\n\t\t\t\t\ttextColor = 'text-white';\n\t\t\t\t\ticonSvg = `<svg class=\"w-5 h-5\" fill=\"currentColor\" viewBox=\"0 0 20 20\">\n\t\t\t\t\t\t<path fill-rule=\"evenodd\" d=\"M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z\" clip-rule=\"evenodd\"></path>\n\t\t\t\t\t</svg>`;\n\t\t\t\t\tbreak;\n\t\t\t\tcase 'error':\n\t\t\t\t\tbgColor = 'bg-red-500';\n\t\t\t\t\ttextColor = 'text-white';\n\t\t\t\t\ticonSvg = `<svg class=\"w-5 h-5\" fill=\"currentColor\" viewBox=\"0 0 20 20\">\n\t\t\t\t\t\t<path fill-rule=\"evenodd\" d=\"M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z\" clip-rule=\"evenodd\"></path>\n\t\t\t\t\t</svg>`;\n\t\t\t\t\tbreak;\n\t\t\t\tcase 'warning':\n\t\t\t\t\tbgColor = 'bg-yellow-500';\n\t\t\t\t\ttextColor = 'text-white';\n\t\t\t\t\ticonSvg = `<svg class=\"w-5 h-5\" fill=\"currentColor\" viewBox=\"0 0 20 20\">\n\t\t\t\t\t\t<path fill-rule=\"evenodd\" d=\"M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z\" clip-rule=\"evenodd\"></path>\n\t\t\t\t\t</svg>`;\n\t\t\t\t\tbreak;\n\t\t\t\tdefault: // info\n\t\t\t\t\tbgColor = 'bg-blue-500';\n\t\t\t\t\ttextColor = 'text-white';\n\t\t\t\t\ticonSvg = `<svg class=\"w-5 h-5\" fill=\"currentColor\" viewBox=\"0 0 20 20\">\n\t\t\t\t\t\t<path fill-rule=\"evenodd\" d=\"M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z\" clip-rule=\"evenodd\"></path>\n\t\t\t\t\t</svg>`;\n\t\t\t}\n\t\t\t\n\t\t\t// 创建通知元素\n\t\t\tconst notification = document.createElement('div');\n\t\t\tnotification.id = notificationId;\n\t\t\tnotification.className = `${bgColor} ${textColor} px-4 py-3 rounded-lg shadow-lg flex items-center space-x-3 min-w-80 max-w-md transform transition-all duration-300 translate-x-full opacity-0`;\n\t\t\tnotification.innerHTML = `\n\t\t\t\t<div class=\"flex-shrink-0\">\n\t\t\t\t\t${iconSvg}\n\t\t\t\t</div>\n\t\t\t\t<div class=\"flex-1\">\n\t\t\t\t\t<p class=\"text-sm font-medium\">${message}</p>\n\t\t\t\t</div>\n\t\t\t\t<button onclick=\"removeNotification('${notificationId}')\" class=\"flex-shrink-0 ml-2 hover:opacity-75 transition-opacity\">\n\t\t\t\t\t<svg class=\"w-4 h-4\" fill=\"currentColor\" viewBox=\"0 0 20 20\">\n\t\t\t\t\t\t<path fill-rule=\"evenodd\" d=\"M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z\" clip-rule=\"evenodd\"></path>\n\t\t\t\t\t</svg>\n\t\t\t\t</button>\n\t\t\t`;\n\t\t\t\n\t\t\t// 添加到容器\n\t\t\tcontainer.appendChild(notification);\n\t\t\t\n\t\t\t// 显示动画\n\t\t\tsetTimeout(() => {\n\t\t\t\tnotification.classList.remove('translate-x-full', 'opacity-0');\n\t\t\t\tnotification.classList.add('translate-x-0', 'opacity-100');\n\t\t\t}, 10);\n\t\t\t\n\t\t\t// 自动移除\n\t\t\tif (duration > 0) {\n\t\t\t\tsetTimeout(() => {\n\t\t\t\t\tremoveNotification(notificationId);\n\t\t\t\t}, duration);\n\t\t\t}\n\t\t}\n\t\t\n\t\t// 移除通知\n\t\tfunction removeNotification(notificationId) {\n\t\t\tconst notification = document.getElementById(notificationId);\n\t\t\tif (notification) {\n\t\t\t\tnotification.classList.remove('translate-x-0', 'opacity-100');\n\t\t\t\tnotification.classList.add('translate-x-full', 'opacity-0');\n\t\t\t\t\n\t\t\t\tsetTimeout(() => {\n\t\t\t\t\tif (notification.parentNode) {\n\t\t\t\t\t\tnotification.parentNode.removeChild(notification);\n\t\t\t\t\t}\n\t\t\t\t}, 300);\n\t\t\t}\n\t\t}\n\t\t\n\t\t// 清除所有通知\n\t\tfunction clearAllNotifications() {\n\t\t\tconst container = document.getElementById('notificationContainer');\n\t\t\tconst notifications = container.querySelectorAll('[id^=\"notification-\"]');\n\t\t\tnotifications.forEach(notification => {\n\t\t\t\tremoveNotification(notification.id);\n\t\t\t});\n\t\t}\n\t</script>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
