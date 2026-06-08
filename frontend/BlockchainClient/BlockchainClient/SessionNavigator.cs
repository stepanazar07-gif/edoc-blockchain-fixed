using System.Linq;
using Microsoft.Maui.Controls;

namespace BlockchainClient
{
    public static class SessionNavigator
    {
        public static void ShowMain()
        {
            var window = Application.Current?.Windows.FirstOrDefault();
            if (window != null)
            {
                window.Page = new AppShell();
            }
        }

        public static void ShowLogin()
        {
            var window = Application.Current?.Windows.FirstOrDefault();
            if (window != null)
            {
                window.Page = new NavigationPage(new LoginPage());
            }
        }
    }
}
