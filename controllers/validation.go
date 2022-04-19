package controllers

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"strconv"

	mailhogv1alpha1 "goimports.patrick.mx/mailhog-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

// ensureCrValid ensures no invalid CRs are processed
func (r *MailhogInstanceReconciler) ensureCrValid(ctx context.Context, cr *mailhogv1alpha1.MailhogInstance) *ReturnIndicator {
	var err error
	name := types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}
	logger := r.logger.WithValues(span, spanCrValid)

	latestCr := &mailhogv1alpha1.MailhogInstance{}
	if err := r.Get(ctx, name, latestCr); err != nil {
		logger.Error(err, failedCrRefresh)
		return &ReturnIndicator{
			Err: err,
		}
	} else if !reflect.DeepEqual(cr, latestCr) {
		logger.Info("cr changed")
		return &ReturnIndicator{}
	}

	checks := []func(*mailhogv1alpha1.MailhogInstance) error{
		checkOverlappingMounts,
		checkMissingSettings,
		checkSmtpUpstreams,
		checkJimFloats,
		checkWebPath,
	}
	for _, check := range checks {
		if err = check(cr); err != nil {
			cr.Status.Error = err.Error()
			if err := r.Status().Update(ctx, cr); err != nil {
				logger.Error(err, failedCrUpdateStatus)
				return &ReturnIndicator{
					Err: err,
				}
			}
			return &ReturnIndicator{
				Err: err,
			}
		}
	}

	logger.Info(stateEnsured)
	return nil
}

// checkOverlappingMounts returns an error if a forbidden mount path is used as maildir path
func checkOverlappingMounts(cr *mailhogv1alpha1.MailhogInstance) error {
	if userPath := cr.Spec.Settings.StorageMaildir.Path; userPath != "" {
		// TODO use "`" for regex to not need escaping
		conflictPathRegex := regexp.MustCompile("^\\/(usr|mailhog)?(\\/)?((settings)\\/?(files)?|(local)\\/?(bin)?\\/?(MailHog)?)?$")
		if matches := conflictPathRegex.MatchString(userPath); matches {
			return errConflictingMount
		}
	}
	return nil
}

// checkMissingSettings returns an error if a storage config is missing detailed settings
func checkMissingSettings(cr *mailhogv1alpha1.MailhogInstance) error {
	if cr.Spec.Settings.Storage == mailhogv1alpha1.MongoDBStorage {
		mongoSpec := cr.Spec.Settings.StorageMongoDb
		if mongoSpec.URI == "" || mongoSpec.Db == "" || mongoSpec.Collection == "" {
			return errMissingMongoDBSettings
		}
	}
	if cr.Spec.Settings.Storage == mailhogv1alpha1.MaildirStorage {
		if cr.Spec.Settings.StorageMaildir.Path == "" {
			return errMissingMaildirSettings
		}
	}

	return nil
}

// checkSmtpUpstreams returns an error if a smtp upstream specifies credentials but no authentication mechanism
func checkSmtpUpstreams(cr *mailhogv1alpha1.MailhogInstance) error {
	if cr.Spec.Settings.Files != nil {
		if len(cr.Spec.Settings.Files.SmtpUpstreams) > 0 {
			for _, upstream := range cr.Spec.Settings.Files.SmtpUpstreams {
				if upstream.Username != "" || upstream.Password != "" {
					if upstream.Mechanism == "" {
						return errMissingUpstreamSmtpMechanism
					}
				}
			}
		}
	}
	return nil
}

// checkJimFloats returns an error if a jim value can not be converted to a float
func checkJimFloats(cr *mailhogv1alpha1.MailhogInstance) error {
	if cr.Spec.Settings.Jim.Invite == true {
		fields := []string{
			cr.Spec.Settings.Jim.RejectRecipient,
			cr.Spec.Settings.Jim.RejectAuth,
			cr.Spec.Settings.Jim.RejectSender,
			cr.Spec.Settings.Jim.Disconnect,
			cr.Spec.Settings.Jim.Accept,
			cr.Spec.Settings.Jim.LinkspeedAffect,
		}
		for _, field := range fields {
			if _, err := strconv.ParseFloat(field, 64); err != nil {
				return errJimNonFloatFound
			}
		}
	}
	return nil
}

// checkWebPath returns an error if the web path begins or ends in a slash
func checkWebPath(cr *mailhogv1alpha1.MailhogInstance) error {
	if path := cr.Spec.Settings.WebPath; path != "" {
		if path[len(path)-1:] == "/" || path[0:1] == "/" {
			return errWebPathNonRelative
		}
	}
	return nil
}

var (
	errConflictingMount             = errors.New("the chosen maildir path conflicts with other paths needed (/usr/local/bin or /mailhog/settings/files)")
	errMissingMongoDBSettings       = errors.New("mongodb was specified as data storage but not all mongodb params have been specified")
	errMissingMaildirSettings       = errors.New("maildir was specified as data storage but no path has been specified")
	errMissingUpstreamSmtpMechanism = errors.New("an upstream smtp server has username / password specified but no auth mechanism")
	errJimNonFloatFound             = errors.New("a chaos monkey probability rate cannot be unpacked as a float")
	errWebPathNonRelative           = errors.New("web path must be relative (not starting or ending with slash)")
)
