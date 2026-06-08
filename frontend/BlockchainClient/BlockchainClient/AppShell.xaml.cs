using Microsoft.Maui.Controls;

namespace BlockchainClient
{
    public partial class AppShell : Shell
    {
        public AppShell()
        {
            InitializeComponent();

            Routing.RegisterRoute(nameof(ProfilePage), typeof(ProfilePage));
            Routing.RegisterRoute(nameof(UploadPage), typeof(UploadPage));
            Routing.RegisterRoute(nameof(MyDocumentsPage), typeof(MyDocumentsPage));
            Routing.RegisterRoute(nameof(MyFilesPage), typeof(MyFilesPage));
            Routing.RegisterRoute(nameof(UsersPage), typeof(UsersPage));
            Routing.RegisterRoute(nameof(SendDocumentPage), typeof(SendDocumentPage));
            Routing.RegisterRoute(nameof(IncomingTransfersPage), typeof(IncomingTransfersPage));
            Routing.RegisterRoute(nameof(DownloadedFilesPage), typeof(DownloadedFilesPage));
            Routing.RegisterRoute(nameof(HistoryPage), typeof(HistoryPage));
        }
    }
}
