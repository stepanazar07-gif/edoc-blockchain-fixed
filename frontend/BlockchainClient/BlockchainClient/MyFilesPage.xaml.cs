using System;
using System.Collections.ObjectModel;
using Microsoft.Maui.Controls;

namespace BlockchainClient
{
    public partial class MyFilesPage : ContentPage
    {
        private readonly ApiService _api = new();

        public ObservableCollection<MyDocument> Files { get; } = new();

        public MyFilesPage()
        {
            InitializeComponent();
            BindingContext = this;
        }

        protected override async void OnAppearing()
        {
            base.OnAppearing();
            await LoadFilesAsync();
        }

        private async void OnRefreshClicked(object sender, EventArgs e)
        {
            await LoadFilesAsync();
        }

        private async System.Threading.Tasks.Task LoadFilesAsync()
        {
            var files = await _api.GetMyDocumentsAsync();
            Files.Clear();
            if (files == null) return;
            foreach (var file in files)
            {
                Files.Add(file);
            }
        }
    }
}
