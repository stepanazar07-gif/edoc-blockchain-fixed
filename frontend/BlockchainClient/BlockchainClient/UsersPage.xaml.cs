using System;
using System.Collections.ObjectModel;
using System.Threading.Tasks;
using Microsoft.Maui.Controls;

namespace BlockchainClient
{
    public partial class UsersPage : ContentPage
    {
        private readonly ApiService _api = new();

        public ObservableCollection<UserInfo> Users { get; } = new();

        public UsersPage()
        {
            InitializeComponent();
            BindingContext = this;
        }

        protected override async void OnAppearing()
        {
            base.OnAppearing();
            await LoadUsersAsync();
        }

        private async void OnSearchTextChanged(object sender, TextChangedEventArgs e)
        {
            if (string.IsNullOrWhiteSpace(e.NewTextValue))
            {
                await LoadUsersAsync();
            }
        }

        private async void OnSearchPressed(object sender, EventArgs e)
        {
            await SearchByIdAsync();
        }

        private async void OnUserSelected(object sender, SelectionChangedEventArgs e)
        {
            if (e.CurrentSelection.Count == 0) return;
            if (e.CurrentSelection[0] is not UserInfo user) return;

            ((CollectionView)sender).SelectedItem = null;
            await Navigation.PushAsync(new SendDocumentPage(user));
        }

        private async Task SearchByIdAsync()
        {
            var id = UserSearchBar.Text?.Trim();
            if (string.IsNullOrWhiteSpace(id))
            {
                await LoadUsersAsync();
                return;
            }

            var list = await _api.GetAllUsersAsync(id);
            ReplaceUsers(list);
        }

        private async Task LoadUsersAsync()
        {
            var list = await _api.GetAllUsersAsync();
            ReplaceUsers(list);
        }

        private void ReplaceUsers(System.Collections.Generic.List<UserInfo>? list)
        {
            Users.Clear();
            if (list == null) return;
            foreach (var user in list)
            {
                Users.Add(user);
            }
        }
    }
}
