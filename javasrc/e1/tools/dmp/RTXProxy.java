package e1.tools.dmp;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.UnsupportedEncodingException;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.URLConnection;
import java.net.URLEncoder;
import java.text.MessageFormat;
import java.util.List;
import java.util.Map;

public class RTXProxy {

	/**
	 * @param args
	 */
	public static void main(String[] args) {
		String msg = args[0];
		String storyID = args[1];
		String receiver = args[2];
		receiver = receiver.replace('-', ',');
		String title = args[3];
		String toURL = null;
		try {
			msg = URLEncoder.encode(msg, "GBK");
			toURL = "192.168.0.10/zentaopms/www/index.php?m=story&f=view&storyID={0}";
			toURL = MessageFormat.format(toURL, storyID);
			toURL = URLEncoder.encode(toURL, "gbk");
		} catch (UnsupportedEncodingException e1) {
			e1.printStackTrace();
		}
		
		Object []params = {msg, toURL, receiver, title};
		String model = "http://192.168.0.12:8012/SendNotify.cgi?msg=[{0}|{1}]&receiver={2}&title={3}";
		if (storyID.equalsIgnoreCase("0")) {
			model = "http://192.168.0.12:8012/SendNotify.cgi?msg={0}&receiver={1}&title={2}";
			params = new Object[]{msg, receiver, title};
		}
		
		String url = MessageFormat.format(model, params);
		System.out.println(url);
		sendGet(url);
	}

	public static String sendGet(String url)  {

		String result = "";
		URL realUrl;
		try {
			realUrl = new URL(url);

			// 打开和URL之间的连接
			URLConnection connection = realUrl.openConnection();
			// 设置通用的请求属性
			connection.setRequestProperty("accept", "*/*");
			connection.setRequestProperty("connection", "Keep-Alive");
			connection
					.setRequestProperty("user-agent",
							"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0");
			// 建立实际的连接
			connection.connect();

			// 获取所有响应头字段
			Map<String, List<String>> map = connection.getHeaderFields();
			// 遍历所有的响应头字段
//			for (String key : map.keySet()) {
//				System.out.println(key + "--->" + map.get(key));
//			}

			// 定义 BufferedReader输入流来读取URL的响应
			BufferedReader in = new BufferedReader(new InputStreamReader(
					connection.getInputStream()));
			String line;
			while ((line = in.readLine()) != null) {
				result += line;
			}

		} catch (MalformedURLException e) {
			e.printStackTrace();
		} catch (IOException e) {
			e.printStackTrace();
		}

		return result;
	}

}
